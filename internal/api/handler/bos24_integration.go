package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/dynamicweblab/masaar-crm/internal/bos24"
	"github.com/dynamicweblab/masaar-crm/internal/config"
	"github.com/dynamicweblab/masaar-crm/internal/domain"
	"github.com/dynamicweblab/masaar-crm/internal/repo"
)

// BOS24IntegrationHandler handles BOS24 marketplace integration:
//
//	POST /webhooks/bos24?token=<secret>              — receive real-time events from BOS24
//	GET  /api/v1/settings/bos24-integration          — fetch integration status
//	PATCH /api/v1/settings/bos24-integration         — save API key
//	POST /api/v1/settings/bos24-integration/register — register webhook with BOS24
//	POST /api/v1/settings/bos24-integration/sync     — manual delta sync
type BOS24IntegrationHandler struct {
	bos24Repo   *repo.BOS24IntegrationRepo
	syncService *bos24.SyncService
	cfg         *config.Config
}

func NewBOS24IntegrationHandler(
	bos24Repo *repo.BOS24IntegrationRepo,
	syncService *bos24.SyncService,
	cfg *config.Config,
) *BOS24IntegrationHandler {
	return &BOS24IntegrationHandler{
		bos24Repo:   bos24Repo,
		syncService: syncService,
		cfg:         cfg,
	}
}

// ── Public webhook receiver ───────────────────────────────────────────────────

// ReceiveWebhook handles POST /webhooks/bos24?token=<secret>
// BOS24 calls this when listing or inquiry events occur.
// The ?token identifies the company; X-BOS24-Webhook-Signature is verified
// with HMAC-SHA256 using the same secret as the token.
func (h *BOS24IntegrationHandler) ReceiveWebhook(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Look up company by webhook secret
	settings, err := h.bos24Repo.GetCompanyByWebhookSecret(c.Context(), token)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Verify HMAC-SHA256 signature — same pattern as WhatsApp webhook handler
	sig := c.Get("X-BOS24-Webhook-Signature")
	if sig == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	mac := hmac.New(sha256.New, []byte(settings.WebhookSecret))
	mac.Write(c.Body())
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Parse the event envelope
	var event bos24.BOS24WebhookEvent
	if err := c.BodyParser(&event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	// Process asynchronously — BOS24 expects a 2xx response within 8 seconds
	go func() {
		if err := h.syncService.ProcessEvent(c.Context(), settings.CompanyID, &event); err != nil {
			log.Printf("[BOS24] webhook event %s for company %s: %v", event.Event, settings.CompanyID, err)
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}

// ── Admin settings endpoints ──────────────────────────────────────────────────

// GetSettings handles GET /api/v1/settings/bos24-integration
func (h *BOS24IntegrationHandler) GetSettings(c *fiber.Ctx) error {
	companyID, err := uuid.Parse(c.Locals("company_id").(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid company_id"})
	}

	settings, err := h.bos24Repo.GetSettings(c.Context(), companyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load settings"})
	}

	webhookURL := ""
	if settings.WebhookSecret != "" {
		webhookURL = h.cfg.AppURL + "/webhooks/bos24?token=" + settings.WebhookSecret
	}

	return c.JSON(fiber.Map{
		"api_key_set":        settings.APIKey != "",
		"api_key_preview":    maskSecret(settings.APIKey, 8),
		"webhook_registered": settings.IsWebhookRegistered(),
		"webhook_url":        webhookURL,
		"webhook_secret":     maskSecret(settings.WebhookSecret, 6),
		"webhook_id":         settings.WebhookID,
		"last_sync_at":       settings.LastSyncAt,
	})
}

// UpdateSettings handles PATCH /api/v1/settings/bos24-integration
// Body: { "api_key": "bos24_live_..." }
// Auto-generates a webhook_secret on first save so the webhook URL is ready.
func (h *BOS24IntegrationHandler) UpdateSettings(c *fiber.Ctx) error {
	companyID, err := uuid.Parse(c.Locals("company_id").(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid company_id"})
	}

	var body struct {
		APIKey string `json:"api_key"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if body.APIKey == "" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "api_key is required"})
	}

	// Load existing settings to preserve webhook_secret and webhook_id
	existing, _ := h.bos24Repo.GetSettings(c.Context(), companyID)
	if existing == nil {
		existing = &domain.BOS24Settings{CompanyID: companyID}
	}
	existing.APIKey = body.APIKey

	// Auto-generate a webhook secret if not already set
	if existing.WebhookSecret == "" {
		secret, err := generateSecret(32)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate webhook secret"})
		}
		existing.WebhookSecret = secret
	}

	if err := h.bos24Repo.SaveSettings(c.Context(), existing); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save settings"})
	}

	webhookURL := h.cfg.AppURL + "/webhooks/bos24?token=" + existing.WebhookSecret
	return c.JSON(fiber.Map{
		"ok":                 true,
		"webhook_url":        webhookURL,
		"webhook_registered": existing.IsWebhookRegistered(),
	})
}

// RegisterWebhook handles POST /api/v1/settings/bos24-integration/register
// Calls BOS24's webhook registration endpoint and saves the returned webhook ID.
func (h *BOS24IntegrationHandler) RegisterWebhook(c *fiber.Ctx) error {
	companyID, err := uuid.Parse(c.Locals("company_id").(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid company_id"})
	}

	settings, err := h.bos24Repo.GetSettings(c.Context(), companyID)
	if err != nil || !settings.IsConfigured() {
		return c.Status(fiber.StatusPreconditionFailed).JSON(fiber.Map{
			"error": "BOS24 API key not configured — save your API key first",
		})
	}
	if settings.WebhookSecret == "" {
		return c.Status(fiber.StatusPreconditionFailed).JSON(fiber.Map{
			"error": "Webhook secret not yet generated — save your API key first",
		})
	}

	webhookURL := h.cfg.AppURL + "/webhooks/bos24?token=" + settings.WebhookSecret
	events := []string{
		"listing.created",
		"listing.updated",
		"listing.published",
		"listing.deactivated",
		"listing.deleted",
		"inquiry.created",
	}

	client := bos24.NewIntegrationClient(settings.APIKey)
	reg, err := client.RegisterWebhook(c.Context(), webhookURL, events, settings.WebhookSecret)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Failed to register webhook with BOS24: " + err.Error(),
		})
	}

	settings.WebhookID = fmt.Sprintf("%d", reg.ID)
	if err := h.bos24Repo.SaveSettings(c.Context(), settings); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save webhook ID"})
	}

	return c.JSON(fiber.Map{
		"ok":          true,
		"webhook_url": webhookURL,
		"webhook_id":  settings.WebhookID,
	})
}

// SyncNow handles POST /api/v1/settings/bos24-integration/sync
// Triggers an immediate delta sync (listings + inquiries since last_sync_at).
func (h *BOS24IntegrationHandler) SyncNow(c *fiber.Ctx) error {
	companyID, err := uuid.Parse(c.Locals("company_id").(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid company_id"})
	}

	settings, err := h.bos24Repo.GetSettings(c.Context(), companyID)
	if err != nil || !settings.IsConfigured() {
		return c.Status(fiber.StatusPreconditionFailed).JSON(fiber.Map{
			"error": "BOS24 API key not configured",
		})
	}

	result, err := h.syncService.SyncAll(c.Context(), companyID, settings.APIKey, settings.LastSyncAt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	now := time.Now()
	_ = h.bos24Repo.UpdateLastSyncAt(c.Context(), companyID, now)

	return c.JSON(fiber.Map{
		"ok":                true,
		"listings_imported": result.ListingsImported,
		"inquiries_created": result.InquiriesCreated,
		"synced_at":         now,
	})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// maskSecret shows the first showFirst chars + "..." + last 4 chars.
// Returns empty string if the input is empty.
func maskSecret(s string, showFirst int) string {
	if s == "" {
		return ""
	}
	if len(s) <= showFirst+4 {
		return s[:1] + "***"
	}
	return s[:showFirst] + "..." + s[len(s)-4:]
}

// generateSecret produces a cryptographically random hex string of byteLen random bytes.
func generateSecret(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
