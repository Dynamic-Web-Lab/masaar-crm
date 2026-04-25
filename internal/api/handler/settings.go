package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maidulcu/masaar-crm/internal/domain"
	"github.com/maidulcu/masaar-crm/internal/repo"
)

type SettingsHandler struct {
	repo *repo.SettingsRepo
}

func NewSettingsHandler(repo *repo.SettingsRepo) *SettingsHandler {
	return &SettingsHandler{repo: repo}
}

// GetBOS24Settings retrieves the BOS24 API token configuration
// @Summary Get BuyOrSell24 API settings
// @Description Retrieve BuyOrSell24 API token (admin only). Used to check if integration is configured.
// @Tags Settings
// @Produce json
// @Success 200 {object} domain.APISetting
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/settings/bos24 [get]
// @Security Bearer
func (h *SettingsHandler) GetBOS24Settings(c *fiber.Ctx) error {
	userRole := c.Locals("role").(domain.Role)
	if userRole != domain.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "admin access required",
		})
	}

	setting, err := h.repo.Get(c.Context(), "bos24_api_token")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve settings",
		})
	}

	return c.JSON(setting)
}

// UpdateBOS24Settings updates the BOS24 API token
// @Summary Update BuyOrSell24 API token
// @Description Update BuyOrSell24 API token (admin only). Token is required to enable real estate features.
// @Tags Settings
// @Accept json
// @Produce json
// @Param request body UpdateBOS24Request true "New API token"
// @Success 200 {object} domain.APISetting
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/settings/bos24 [patch]
// @Security Bearer
func (h *SettingsHandler) UpdateBOS24Settings(c *fiber.Ctx) error {
	userRole := c.Locals("role").(domain.Role)
	if userRole != domain.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "admin access required",
		})
	}

	var req UpdateBOS24Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	userID := c.Locals("user_id").(uuid.UUID)

	if err := h.repo.UpdateBOS24Token(c.Context(), req.Token, &userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update settings",
		})
	}

	updated, err := h.repo.Get(c.Context(), "bos24_api_token")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve updated settings",
		})
	}

	return c.JSON(updated)
}

// Request types
type UpdateBOS24Request struct {
	Token string `json:"token"`
}
