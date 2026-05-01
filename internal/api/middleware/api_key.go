package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/maidulcu/masaar-crm/internal/repo"
)

// ValidateAPIKey checks Authorization header for "Bearer sk_live_..." format.
// On success, stores api_key_id, company_id, api_key_scopes in Fiber locals.
func ValidateAPIKey(apiKeyRepo *repo.ApiKeyRepo) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
		}

		// Parse "Bearer sk_live_..."
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization format"})
		}

		plaintext := parts[1]
		if !strings.HasPrefix(plaintext, "sk_live_") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid api key format"})
		}

		// Hash the key
		hashSum := sha256.Sum256([]byte(plaintext))
		keyHash := hex.EncodeToString(hashSum[:])

		// Validate against database
		apiKey, err := apiKeyRepo.ValidateKey(c.Context(), keyHash)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or revoked api key"})
		}

		// Store in locals
		c.Locals("api_key_id", apiKey.ID)
		c.Locals("company_id", apiKey.CompanyID)
		c.Locals("api_key_scopes", apiKey.Scopes)
		c.Locals("api_key_name", apiKey.Name)

		return c.Next()
	}
}

// RequireAPIKeyScope checks if the API key has a specific scope.
func RequireAPIKeyScope(requiredScope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		scopes := c.Locals("api_key_scopes")
		if scopes == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "missing api key scopes"})
		}

		scopesStr := scopes.(string)
		for _, scope := range strings.Split(scopesStr, ",") {
			if strings.TrimSpace(scope) == requiredScope {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "insufficient permissions"})
	}
}
