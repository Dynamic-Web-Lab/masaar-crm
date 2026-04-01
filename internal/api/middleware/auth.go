package middleware

import (
	"strings"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/maidulcu/masaar-crm/internal/domain"
	"github.com/redis/go-redis/v9"
)

func CheckBlacklist(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenVal := c.Locals("user")
		if tokenVal != nil {
			if token, ok := tokenVal.(*jwt.Token); ok {
				val, err := rdb.Get(c.UserContext(), "blacklist:"+token.Raw).Result()
				if err == nil && val == "1" {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
						"error": "token has been revoked",
					})
				}
			}
		}
		return c.Next()
	}
}

func JWT(secret string) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(secret)},
		ErrorHandler: jwtError,
	})
}

func jwtError(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": "unauthorized",
	})
}

// RequireRole returns 403 if the authenticated user doesn't have one of the given roles.
func RequireRole(roles ...domain.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)
		roleVal, ok := claims["role"]
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "role claim missing",
			})
		}
		role := domain.Role(roleVal.(string))
		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden",
		})
	}
}

// ClaimsFromCtx extracts JWT claims from Fiber context.
func ClaimsFromCtx(c *fiber.Ctx) jwt.MapClaims {
	return c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)
}

// BearerToken extracts the raw token string from Authorization header.
func BearerToken(c *fiber.Ctx) string {
	auth := c.Get("Authorization")
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1]
	}
	return ""
}
