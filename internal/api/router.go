package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberws "github.com/gofiber/websocket/v2"
	_ "github.com/maidulcu/masaar-crm/docs" // swagger generated docs
	"github.com/maidulcu/masaar-crm/internal/api/handler"
	"github.com/maidulcu/masaar-crm/internal/api/middleware"
	"github.com/maidulcu/masaar-crm/internal/config"
	"github.com/maidulcu/masaar-crm/internal/domain"
	"github.com/maidulcu/masaar-crm/internal/ws"
	"github.com/redis/go-redis/v9"
	fiberswagger "github.com/swaggo/fiber-swagger"
)

type Handlers struct {
	Auth         *handler.AuthHandler
	User         *handler.UserHandler
	Stats        *handler.StatsHandler
	Contact      *handler.ContactHandler
	Lead         *handler.LeadHandler
	WhatsApp     *handler.WhatsAppHandler
	AI           *handler.AIHandler
	Message      *handler.MessageHandler
	Notification *handler.NotificationHandler
	Deal         *handler.DealHandler
	Invoice      *handler.InvoiceHandler
	Property     *handler.PropertyHandler
	Settings     *handler.SettingsHandler
	Email        *handler.EmailHandler
}

// webhookLimiter allows Meta's burst delivery (300 req/min per IP) while
// blocking abuse. Meta retries on 429 so legitimate messages are never lost.
var webhookLimiter = limiter.New(limiter.Config{
	Max:        300,
	Expiration: 1 * time.Minute,
	LimitReached: func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "rate limit exceeded"})
	},
})

// loginLimiter prevents brute-force on the auth endpoint.
var loginLimiter = limiter.New(limiter.Config{
	Max:        10,
	Expiration: 1 * time.Minute,
	LimitReached: func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many login attempts"})
	},
})

func RegisterRoutes(app *fiber.App, h *Handlers, hub *ws.Hub, cfg *config.Config, rdb *redis.Client) {
	// ── Public routes ────────────────────────────────────────────────────────
	app.Post("/api/v1/auth/login", loginLimiter, h.Auth.Login)
	app.Post("/api/v1/auth/refresh", h.Auth.Refresh)

	// WhatsApp webhook — Meta calls this publicly
	app.Get("/webhooks/whatsapp", h.WhatsApp.Verify)
	app.Post("/webhooks/whatsapp", webhookLimiter, h.WhatsApp.Receive)

	// ── WebSocket — authenticated upgrade ────────────────────────────────────
	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Personal notifications
	app.Get("/ws/notifications", middleware.JWT(cfg.JWTSecret), middleware.CheckBlacklist(rdb), fiberws.New(hub.Handler()))

	// ── Authenticated API ────────────────────────────────────────────────────
	v1 := app.Group("/api/v1", middleware.JWT(cfg.JWTSecret), middleware.CheckBlacklist(rdb))

	v1.Delete("/auth/logout", h.Auth.Logout)

	// Dashboard stats — all authenticated users
	v1.Get("/stats", h.Stats.Overview)

	// User settings — personal; no role restriction beyond auth
	v1.Get("/users/me", h.User.GetMe)
	v1.Patch("/users/me/password", h.User.ChangePassword)
	v1.Patch("/users/me/lang", h.User.UpdateLang)

	// API Settings — admin only (integrations, API keys)
	v1.Get("/settings/bos24",
		middleware.RequireRole(domain.RoleAdmin),
		h.Settings.GetBOS24Settings,
	)
	v1.Patch("/settings/bos24",
		middleware.RequireRole(domain.RoleAdmin),
		h.Settings.UpdateBOS24Settings,
	)

	// Company Settings — admin only (invoice details, VAT number, bank info)
	v1.Get("/settings/company",
		middleware.RequireRole(domain.RoleAdmin),
		h.Settings.GetCompanySettings,
	)
	v1.Patch("/settings/company",
		middleware.RequireRole(domain.RoleAdmin),
		h.Settings.UpdateCompanySettings,
	)

	// Contacts — viewers: read-only; agents: create+update; admin: delete
	v1.Get("/contacts", h.Contact.List)
	v1.Get("/contacts/:id", h.Contact.Get)
	v1.Post("/contacts",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Contact.Create,
	)
	v1.Patch("/contacts/:id",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Contact.Update,
	)
	v1.Delete("/contacts/:id",
		middleware.RequireRole(domain.RoleAdmin),
		h.Contact.Delete,
	)

	// Leads / Pipeline — viewers: read-only; agents: create+move; admin: all
	v1.Get("/leads", h.Lead.KanbanBoard)
	v1.Get("/leads/:id", h.Lead.Get)
	v1.Post("/leads",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Lead.Create,
	)
	v1.Patch("/leads/:id/stage",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Lead.UpdateStage,
	)
	v1.Patch("/leads/:id/notes",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Lead.UpdateNotes,
	)

	// WhatsApp inbox — all authenticated users read; agents+ can close
	v1.Get("/threads", h.WhatsApp.ListThreads)
	v1.Get("/threads/:id", h.WhatsApp.GetThread)
	v1.Get("/threads/:id/messages", h.WhatsApp.GetMessages)
	v1.Post("/threads/:id/close",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.WhatsApp.CloseThread,
	)

	// AI (manual) — agents and admin only
	v1.Post("/ai/summarize/:thread_id",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.AI.SummarizeThread,
	)

	// Message Analysis — agents and admin only (intent parsing, enrichment, auto-lead)
	v1.Post("/messages/analyze",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Message.AnalyzeMessage,
	)
	v1.Post("/messages/suggest-action",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Message.SuggestNextAction,
	)
	v1.Post("/messages/auto-create-lead",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Message.AutoCreateLead,
	)

	// Real Estate Market Data (BuyOrSell24) — agents and admin only (optional integration)
	v1.Post("/properties/search",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Property.SearchProperties,
	)
	v1.Get("/properties/transactions",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Property.GetTransactions,
	)
	v1.Get("/properties/buildings",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Property.GetBuildings,
	)
	v1.Get("/properties/buildings/:id",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Property.GetBuildingByID,
	)
	v1.Get("/properties/schools/nearby",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Property.GetNearbySchools,
	)
	v1.Get("/properties/yield-analysis",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Property.GetYieldAnalysis,
	)
	v1.Get("/properties/comparables",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Property.GetComparables,
	)
	v1.Get("/properties/market-trends",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Property.GetMarketTrends,
	)

	// Notifications — personal; no role restriction beyond auth
	v1.Get("/notifications", h.Notification.List)
	v1.Patch("/notifications/:id/read", h.Notification.MarkRead)

	// Deals — viewers: read-only; agents: create+stage; admin: all
	v1.Get("/deals", h.Deal.List)
	v1.Get("/deals/:id/invoices", h.Deal.ListInvoices)
	v1.Post("/deals",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Deal.Create,
	)
	v1.Patch("/deals/:id/stage",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Deal.UpdateStage,
	)

	// Invoices — agents: create+view; admin: send+update status
	v1.Get("/invoices/:id", h.Invoice.Get)
	v1.Get("/invoices/:id/pdf", h.Invoice.DownloadPDF)
	v1.Post("/invoices",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Invoice.Create,
	)
	v1.Post("/invoices/:id/send",
		middleware.RequireRole(domain.RoleAdmin),
		h.Invoice.Send,
	)
	v1.Patch("/invoices/:id/status",
		middleware.RequireRole(domain.RoleAdmin),
		h.Invoice.UpdateStatus,
	)

	// Email — agents and admin only
	v1.Post("/emails/send",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Email.SendEmail,
	)
	v1.Get("/emails/history",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Email.GetEmailHistory,
	)

	// Health
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Swagger UI — available in all envs; gate with BasicAuth in production if needed
	app.Get("/docs/*", fiberswagger.WrapHandler)
}
