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
	Auth              *handler.AuthHandler
	User              *handler.UserHandler
	Stats             *handler.StatsHandler
	Contact           *handler.ContactHandler
	Lead              *handler.LeadHandler
	WhatsApp          *handler.WhatsAppHandler
	WhatsAppOutbound  *handler.WhatsAppOutboundHandler
	AI                *handler.AIHandler
	Message           *handler.MessageHandler
	Notification      *handler.NotificationHandler
	Deal              *handler.DealHandler
	Invoice           *handler.InvoiceHandler
	Property          *handler.PropertyHandler
	Settings          *handler.SettingsHandler
	Email             *handler.EmailHandler
	RentalProperty    *handler.RentalPropertyHandler
	Tenant            *handler.TenantHandler
	LeaseTemplate     *handler.LeaseTemplateHandler
	Lease             *handler.LeaseHandler
	Payment           *handler.PaymentHandler
	BankIntegration   *handler.BankIntegrationHandler
	BankStatement     *handler.BankStatementHandler
	PaymentConfirmation *handler.PaymentConfirmationHandler
	Analytics         *handler.AnalyticsHandler
	Expense           *handler.ExpenseHandler
	Inspection        *handler.InspectionHandler
	Maintenance       *handler.MaintenanceTaskHandler
	LeaseRenewal      *handler.LeaseRenewalHandler
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

// apiLimiter caps general authenticated API usage to 100 requests/min per IP.
var apiLimiter = limiter.New(limiter.Config{
	Max:        100,
	Expiration: 1 * time.Minute,
	LimitReached: func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "rate limit exceeded"})
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
	app.Get("/ws/notifications",
		middleware.JWT(cfg.JWTSecret),
		middleware.CheckBlacklist(rdb),
		middleware.ExtractClaims(cfg.CompanyID),
		fiberws.New(hub.Handler()),
	)

	// ── Authenticated API ────────────────────────────────────────────────────
	v1 := app.Group("/api/v1",
		apiLimiter,
		middleware.JWT(cfg.JWTSecret),
		middleware.CheckBlacklist(rdb),
		middleware.ExtractClaims(cfg.CompanyID),
	)

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
	v1.Get("/leads/:id/communications", h.Lead.GetCommunications)

	// WhatsApp inbox — all authenticated users read; agents+ can close/send
	v1.Get("/threads", h.WhatsApp.ListThreads)
	v1.Get("/threads/:id", h.WhatsApp.GetThread)
	v1.Get("/threads/:id/messages", h.WhatsApp.GetMessages)
	v1.Post("/threads/:id/close",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.WhatsApp.CloseThread,
	)

	// WhatsApp outbound — agents+ send messages
	v1.Post("/threads/:id/send-message",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.WhatsAppOutbound.SendMessage,
	)
	v1.Post("/threads/:id/send-template",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.WhatsAppOutbound.SendTemplate,
	)
	v1.Get("/threads/:id/outbound-messages",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.WhatsAppOutbound.GetOutboundMessages,
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

	// Rental Properties — agents: create+view+update; admin: all; viewers: read-only
	v1.Get("/rental-properties", h.RentalProperty.List)
	v1.Get("/rental-properties/:id", h.RentalProperty.Get)
	v1.Post("/rental-properties",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.RentalProperty.Create,
	)
	v1.Patch("/rental-properties/:id",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.RentalProperty.Update,
	)
	v1.Delete("/rental-properties/:id",
		middleware.RequireRole(domain.RoleAdmin),
		h.RentalProperty.Delete,
	)

	// Tenants — agents: create+view+update; admin: all; viewers: read-only
	v1.Get("/tenants", h.Tenant.List)
	v1.Get("/tenants/:id", h.Tenant.Get)
	v1.Post("/tenants",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Tenant.Create,
	)
	v1.Patch("/tenants/:id",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Tenant.Update,
	)
	v1.Delete("/tenants/:id",
		middleware.RequireRole(domain.RoleAdmin),
		h.Tenant.Delete,
	)
	v1.Post("/tenants/:id/verify",
		middleware.RequireRole(domain.RoleAdmin),
		h.Tenant.Verify,
	)

	// Lease Templates — agents: view; admin: all
	v1.Get("/lease-templates", h.LeaseTemplate.List)
	v1.Get("/lease-templates/:id", h.LeaseTemplate.Get)
	v1.Post("/lease-templates",
		middleware.RequireRole(domain.RoleAdmin),
		h.LeaseTemplate.Create,
	)
	v1.Patch("/lease-templates/:id",
		middleware.RequireRole(domain.RoleAdmin),
		h.LeaseTemplate.Update,
	)
	v1.Delete("/lease-templates/:id",
		middleware.RequireRole(domain.RoleAdmin),
		h.LeaseTemplate.Delete,
	)

	// Leases — agents: create+view+update; admin: all; viewers: read-only
	v1.Get("/leases", h.Lease.List)
	v1.Get("/leases/:id", h.Lease.Get)
	v1.Post("/leases",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Lease.Create,
	)
	v1.Patch("/leases/:id",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Lease.Update,
	)
	v1.Delete("/leases/:id",
		middleware.RequireRole(domain.RoleAdmin),
		h.Lease.Delete,
	)

	// Payments — agents: create+view+update; admin: all; viewers: read-only
	v1.Get("/payments", h.Payment.List)
	v1.Get("/payments/:id", h.Payment.Get)
	v1.Post("/payments",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Payment.Create,
	)
	v1.Patch("/payments/:id",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Payment.Update,
	)
	v1.Delete("/payments/:id",
		middleware.RequireRole(domain.RoleAdmin),
		h.Payment.Delete,
	)

	// Bank Integrations — admin only
	v1.Get("/bank-integrations", h.BankIntegration.List)
	v1.Get("/bank-integrations/:id", h.BankIntegration.Get)
	v1.Post("/bank-integrations",
		middleware.RequireRole(domain.RoleAdmin),
		h.BankIntegration.Create,
	)
	v1.Patch("/bank-integrations/:id",
		middleware.RequireRole(domain.RoleAdmin),
		h.BankIntegration.Update,
	)
	v1.Delete("/bank-integrations/:id",
		middleware.RequireRole(domain.RoleAdmin),
		h.BankIntegration.Delete,
	)

	// Bank Statements — agents: upload+view; admin: all
	v1.Get("/bank-statements",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.BankStatement.List,
	)
	v1.Get("/bank-statements/:id",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.BankStatement.Get,
	)
	v1.Post("/bank-statements/upload",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.BankStatement.Upload,
	)
	v1.Delete("/bank-statements/:id",
		middleware.RequireRole(domain.RoleAdmin),
		h.BankStatement.Delete,
	)

	// Payment Confirmations — agents: view+send; admin: all
	v1.Get("/payments/:payment_id/confirmation",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.PaymentConfirmation.GetByPayment,
	)
	v1.Post("/payments/:payment_id/send-confirmation",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.PaymentConfirmation.Send,
	)

	// Analytics — read-only; all authenticated users
	v1.Get("/analytics/tenant-overview", h.Analytics.GetTenantOverview)
	v1.Get("/analytics/properties", h.Analytics.ListPropertiesAnalytics)
	v1.Get("/analytics/properties/:propertyID", h.Analytics.GetPropertyAnalytics)
	v1.Get("/analytics/tenants", h.Analytics.ListTenantsPerformance)
	v1.Get("/analytics/tenants/:tenantID", h.Analytics.GetTenantPerformance)
	v1.Get("/analytics/financial", h.Analytics.GetFinancialAnalytics)
	v1.Get("/analytics/maintenance", h.Analytics.GetMaintenanceAnalytics)

	// Expenses — agents: create+view+update; admin: all
	v1.Get("/expense-categories",
		h.Expense.ListCategories,
	)
	v1.Post("/expense-categories",
		middleware.RequireRole(domain.RoleAdmin),
		h.Expense.CreateCategory,
	)
	v1.Get("/expenses", h.Expense.ListExpenses)
	v1.Get("/expenses/:id", h.Expense.GetExpense)
	v1.Post("/expenses",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Expense.CreateExpense,
	)
	v1.Patch("/expenses/:id",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Expense.UpdateExpense,
	)
	v1.Delete("/expenses/:id",
		middleware.RequireRole(domain.RoleAdmin),
		h.Expense.DeleteExpense,
	)
	v1.Post("/expenses/:id/approve",
		middleware.RequireRole(domain.RoleAdmin),
		h.Expense.ApproveExpense,
	)

	// Inspection Templates — admin: create/update; all: list
	v1.Get("/inspection-templates", h.Inspection.ListTemplates)
	v1.Post("/inspection-templates",
		middleware.RequireRole(domain.RoleAdmin),
		h.Inspection.CreateTemplate,
	)

	// Inspections — agents: create+view+update; admin: all
	v1.Get("/inspections", h.Inspection.ListInspections)
	v1.Get("/inspections/:id", h.Inspection.GetInspection)
	v1.Post("/inspections",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Inspection.CreateInspection,
	)
	v1.Patch("/inspections/:id",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Inspection.UpdateInspection,
	)
	v1.Post("/inspections/:id/complete",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Inspection.CompleteInspection,
	)

	// Maintenance Tasks — agents: create+view+update; admin: all
	v1.Get("/maintenance-tasks", h.Maintenance.List)
	v1.Get("/maintenance-tasks/:id", h.Maintenance.Get)
	v1.Post("/maintenance-tasks",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Maintenance.Create,
	)
	v1.Patch("/maintenance-tasks/:id",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Maintenance.Update,
	)
	v1.Post("/maintenance-tasks/:id/complete",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Maintenance.Complete,
	)
	v1.Post("/maintenance-tasks/:id/photos",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.Maintenance.AddPhoto,
	)
	v1.Get("/maintenance-tasks/:id/photos", h.Maintenance.GetPhotos)
	v1.Delete("/maintenance-tasks/:id",
		middleware.RequireRole(domain.RoleAdmin),
		h.Maintenance.Delete,
	)

	// Lease Renewals
	v1.Get("/lease-renewals", h.LeaseRenewal.List)
	v1.Get("/lease-renewals/:id", h.LeaseRenewal.Get)
	v1.Post("/lease-renewals/:lease_id/initiate",
		middleware.RequireRole(domain.RoleAdmin),
		h.LeaseRenewal.Initiate,
	)
	v1.Put("/lease-renewals/:id/propose",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.LeaseRenewal.Propose,
	)
	v1.Post("/lease-renewals/:id/send-offer",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.LeaseRenewal.SendOffer,
	)
	v1.Put("/lease-renewals/:id/accept",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.LeaseRenewal.Accept,
	)
	v1.Put("/lease-renewals/:id/reject",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.LeaseRenewal.Reject,
	)
	v1.Post("/lease-renewals/:id/counter-offer",
		middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent),
		h.LeaseRenewal.CounterOffer,
	)

	// Renewal Templates
	v1.Get("/renewal-templates", h.LeaseRenewal.ListTemplates)
	v1.Post("/renewal-templates",
		middleware.RequireRole(domain.RoleAdmin),
		h.LeaseRenewal.CreateTemplate,
	)

	// Health
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Swagger UI — available in all envs; gate with BasicAuth in production if needed
	app.Get("/docs/*", fiberswagger.WrapHandler)
}
