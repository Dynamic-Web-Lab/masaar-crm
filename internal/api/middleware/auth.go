package middleware

import (
	"strings"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/maidulcu/masaar-crm/internal/domain"
	"github.com/redis/go-redis/v9"
)

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

// BlockDemoWrites rejects all state-changing requests for tokens that carry
// "demo": true in their claims. GET/HEAD/OPTIONS pass through; DELETE on the
// logout path is also allowed so demo users can still sign out cleanly.
// This is belt-and-suspenders on top of the Viewer RBAC — no demo user can
// ever send emails, spam WhatsApp, mutate data, or consume AI quota.
func BlockDemoWrites() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		demo, _ := claims["demo"].(bool)
		if !demo {
			return c.Next()
		}

		method := c.Method()
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return c.Next()
		}
		// Allow the demo user to log out
		if method == "DELETE" && strings.HasSuffix(c.Path(), "/auth/logout") {
			return c.Next()
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "demo accounts are read-only",
			"demo":  true,
		})
	}
}

// ExtractClaims reads JWT claims and sets user_id (uuid.UUID), company_id (string),
// and role (domain.Role) in Fiber locals so handlers can access them without
// repeating assertion boilerplate.
func ExtractClaims(companyID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		sub, _ := claims["sub"].(string)
		userID, err := uuid.Parse(sub)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token subject"})
		}
		c.Locals("user_id", userID)
		c.Locals("company_id", companyID)
		if roleStr, ok := claims["role"].(string); ok {
			c.Locals("role", domain.Role(roleStr))
		}
		return c.Next()
	}
}

// CheckBlacklist verifies that the current access token has not been revoked.
// If Redis fails, it fails closed returning 500.
func CheckBlacklist(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := BearerToken(c)
		if token != "" {
			exists, err := rdb.Exists(c.Context(), "blacklist:"+token).Result()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "internal server error",
				})
			}
			if exists > 0 {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "token revoked",
				})
			}
		}
		return c.Next()
	}
}
