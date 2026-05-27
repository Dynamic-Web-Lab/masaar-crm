package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/maidulcu/masaar-crm/internal/api/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/maidulcu/masaar-crm/internal/ai"
	"github.com/maidulcu/masaar-crm/internal/api"
	"github.com/maidulcu/masaar-crm/internal/api/handler"
	"github.com/maidulcu/masaar-crm/internal/billing"
	"github.com/maidulcu/masaar-crm/internal/bos24"
	"github.com/maidulcu/masaar-crm/internal/config"
	"github.com/maidulcu/masaar-crm/internal/email"
	"github.com/maidulcu/masaar-crm/internal/repo"
	"github.com/maidulcu/masaar-crm/internal/sms"
	"github.com/maidulcu/masaar-crm/internal/webhook"
	"github.com/maidulcu/masaar-crm/internal/whatsapp"
	"github.com/maidulcu/masaar-crm/internal/ws"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	// ── Startup validation ───────────────────────────────────────────────────
	if cfg.JWTSecret == "change-me-in-production" || cfg.JWTSecret == "" {
		if cfg.AppEnv == "production" {
			log.Fatal("JWT_SECRET must be set to a strong random value in production")
		}
		log.Println("WARNING: JWT_SECRET is using default value — change before deploying to production")
	}
	if cfg.AppEnv == "production" && cfg.AllowedOrigins == "*" {
		log.Fatal("ALLOWED_ORIGINS must not be '*' in production — set it to your frontend domain")
	}
	if cfg.WAAppSecret == "" && cfg.WAPhoneNumberID != "" {
		// WhatsApp is actively configured (phone number ID set) but App Secret is missing
		if cfg.AppEnv == "production" {
			log.Fatal("WA_APP_SECRET must be set when WhatsApp (WA_PHONE_NUMBER_ID) is configured")
		}
		log.Println("WARNING: WA_APP_SECRET is not set — Meta webhook signature validation disabled")
	}

	// ── Database ─────────────────────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := repo.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	// ── Run migrations ───────────────────────────────────────────────────────
	sqlDB, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open sql db: %v", err)
	}
	goose.SetBaseFS(os.DirFS("migrations"))
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose dialect: %v", err)
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		log.Fatalf("goose up: %v", err)
	}
	sqlDB.Close()

	// ── Redis ────────────────────────────────────────────────────────────────
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis url: %v", err)
	}
	rdb := redis.NewClient(redisOpts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis ping: %v", err)
	}
	defer rdb.Close()

	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo := repo.NewUserRepo(pool)
	contactRepo := repo.NewContactRepo(pool)
	leadRepo := repo.NewLeadRepo(pool)
	waRepo := repo.NewWhatsAppRepo(pool)
	notificationRepo := repo.NewNotificationRepo(pool)
	dealRepo := repo.NewDealRepo(pool)
	invoiceRepo := repo.NewInvoiceRepo(pool)
	statsRepo := repo.NewStatsRepo(pool)
	settingsRepo := repo.NewSettingsRepo(pool)
	companySettingsRepo := repo.NewCompanySettingsRepo(pool)
	emailRepo := repo.NewEmailRepository(pool)
	outboundRepo := repo.NewWhatsAppOutboundRepo(pool)
	leadTagRepo := repo.NewLeadTagRepo(pool)
	commHistRepo := repo.NewCommunicationHistoryRepo(pool)
	rentalPropertyRepo := repo.NewRentalPropertyRepo(pool)
	tenantRepo := repo.NewTenantRepo(pool)
	leaseTemplateRepo := repo.NewLeaseTemplateRepo(pool)
	leaseRepo := repo.NewLeaseRepo(pool)
	paymentRepo := repo.NewPaymentRepo(pool)
	bankIntegrationRepo := repo.NewBankIntegrationRepo(pool)
	paymentReminderRepo := repo.NewPaymentReminderRepo(pool)
	bankStatementRepo := repo.NewBankStatementRepo(pool)
	paymentConfirmationRepo := repo.NewPaymentConfirmationRepo(pool)
	expenseRepo := repo.NewExpenseRepository(pool)
	inspectionTemplateRepo := repo.NewInspectionTemplateRepo(pool)
	inspectionRepo := repo.NewInspectionRepo(pool)
	maintenanceRepo := repo.NewMaintenanceTaskRepo(pool)
	leaseRenewalRepo := repo.NewLeaseRenewalRepo(pool)
	renewalTemplateRepo := repo.NewRenewalTemplateRepo(pool)
	renewalCommLogRepo := repo.NewRenewalCommunicationLogRepo(pool)
	companyRepo := repo.NewCompanyRepo(pool)
	documentRepo := repo.NewDocumentRepo(pool)
	messageTemplateRepo := repo.NewMessageTemplateRepo(pool)

	// ── Email service (SMTP or Azure Communication Services) ─────────────────
	emailService := email.NewService(&email.Config{
		SMTPHost:        cfg.SMTPHost,
		SMTPPort:        cfg.SMTPPort,
		SMTPUser:        cfg.SMTPUser,
		SMTPPassword:    cfg.SMTPPassword,
		FromEmail:       cfg.SMTPFromEmail,
		FromName:        cfg.SMTPFromName,
		AzureEndpoint:   cfg.AzureCommEndpoint,
		AzureKey:        cfg.AzureCommKey,
		AzureFromAddress: cfg.AzureCommFromAddress,
	})
	if emailService.IsConfigured() {
		log.Printf("Email provider: %s", emailService.ProviderName())
	}

	// ── WhatsApp Sender (optional outbound messaging) ─────────────────────────
	var whatsappSender *whatsapp.Sender
	if cfg.WAPhoneNumberID != "" && cfg.WAAccessToken != "" {
		whatsappSender = whatsapp.NewSender(&whatsapp.SenderConfig{
			BaseURL:       cfg.WABaseURL,
			PhoneNumberID: cfg.WAPhoneNumberID,
			AccessToken:   cfg.WAAccessToken,
		})
		log.Println("WhatsApp outbound messaging enabled")
	}

	// ── WebSocket hub ────────────────────────────────────────────────────────
	hub := ws.NewHub()

	// ── AI clients — dual-provider routing ───────────────────────────────────
	//
	// sensitiveAI (Ollama, local): ALWAYS used for anything touching customer
	// data — WhatsApp messages, contacts, leads, payment details. Data never
	// leaves the server. Required for UAE PDPL compliance.
	//
	// cloudAI (Gemini): ONLY used for non-PII tasks — property listing copy,
	// market summaries, public content. Falls back to sensitiveAI if not set.
	sensitiveAI := ai.NewClient(cfg.OllamaBaseURL, cfg.OllamaModel)
	log.Printf("AI sensitive (local): Ollama %s @ %s", cfg.OllamaModel, cfg.OllamaBaseURL)

	var cloudAI *ai.Client
	if cfg.GeminiAPIKey != "" {
		cloudAI = ai.NewGeminiClient(cfg.GeminiAPIKey, cfg.GeminiModel)
		log.Printf("AI cloud (non-PII): Gemini %s", cfg.GeminiModel)
	} else {
		log.Println("AI cloud: not configured — Ollama handles all tasks")
	}

	// cloudOrLocal returns cloudAI when available, sensitiveAI otherwise.
	cloudOrLocal := func() *ai.Client {
		if cloudAI != nil {
			return cloudAI
		}
		return sensitiveAI
	}

	// ── Scoring service for automatic lead scoring ───────────────────────────
	scoringService := ai.NewScoringService(leadRepo, commHistRepo, leadTagRepo)

	// ── Tagging service for auto-tagging on messages ────────────────────────
	// Uses sensitiveAI — parses raw customer WhatsApp messages (PII)
	taggingService := ai.NewTaggingService(sensitiveAI, leadRepo, waRepo, leadTagRepo, contactRepo)

	// ── SMSCountry client (optional SMS OTP login) ───────────────────────────
	var smsClient *sms.Client
	if sms.IsEnabled(cfg.SMSCountryAuthKey, cfg.SMSCountryAuthToken) {
		smsClient = sms.NewClient(cfg.SMSCountryAuthKey, cfg.SMSCountryAuthToken, cfg.SMSCountrySenderID)
		log.Println("SMSCountry integration enabled")
	}

	// ── BuyOrSell24 client (optional real estate integration) ─────────────────
	var bos24Client *bos24.Client
	// Try to load token from database first, fall back to .env
	dbToken, err := settingsRepo.GetBOS24Token(context.Background())
	if err != nil {
		log.Println("no BOS24 token in database, checking .env")
		dbToken = cfg.BOS24Token
	}
	if bos24.IsEnabled(dbToken) {
		bos24Client = bos24.NewClient(dbToken, cfg.BOS24BaseURL, rdb)
		log.Println("BuyOrSell24 integration enabled")
	}

	auditLogRepo := repo.NewAuditLogRepo(pool)
	apiKeyRepo := repo.NewApiKeyRepo(pool)
	billingRepo := repo.NewBillingRepo(pool)

	// ── Stripe billing (optional) ─────────────────────────────────────────────
	stripeCfg := &billing.StripeConfig{
		SecretKey:       cfg.StripeSecretKey,
		WebhookSecret:   cfg.StripeWebhookSecret,
		PriceIDStarter:  cfg.StripePriceIDStarter,
		PriceIDPro:      cfg.StripePriceIDPro,
		PriceIDBusiness: cfg.StripePriceIDBusiness,
		AppURL:          cfg.AppURL,
	}
	billing.SetupStripe(stripeCfg)
	if stripeCfg.IsEnabled() {
		log.Println("Stripe billing enabled")
	}
	webhookRepo := repo.NewWebhookRepo(pool)
	dispatcher := webhook.NewDispatcher(webhookRepo)

	// ── Payment Reminder Service ──────────────────────────────────────────────
	paymentReminderService := ai.NewPaymentReminderService(
		paymentRepo, paymentReminderRepo, leaseRepo, tenantRepo, contactRepo,
		emailService, whatsappSender, hub,
	)

	// ── Payment Confirmation Service ───────────────────────────────────────────
	paymentConfirmationService := ai.NewPaymentConfirmationService(
		paymentRepo, paymentConfirmationRepo, leaseRepo, tenantRepo, rentalPropertyRepo,
		companySettingsRepo, emailService,
	)

	// ── Handlers ─────────────────────────────────────────────────────────────
	handlers := &api.Handlers{
		Auth:                handler.NewAuthHandler(userRepo, companyRepo, rdb, cfg, auditLogRepo, emailService, smsClient),
		User:                handler.NewUserHandler(userRepo, auditLogRepo, emailService, cfg),
		Stats:               handler.NewStatsHandler(statsRepo),
		Contact:             handler.NewContactHandler(contactRepo, auditLogRepo),
		Lead:                handler.NewLeadHandler(leadRepo, contactRepo, commHistRepo, scoringService, leadTagRepo, hub, auditLogRepo, dispatcher),
		WhatsApp:            handler.NewWhatsAppHandler(waRepo, contactRepo, taggingService, hub, cfg),
		WhatsAppOutbound:    handler.NewWhatsAppOutboundHandler(whatsappSender, outboundRepo, waRepo),
		AI:                  handler.NewAIHandler(sensitiveAI, cloudOrLocal(), contactRepo, leadRepo, waRepo),
		Message:             handler.NewMessageHandler(sensitiveAI, waRepo, contactRepo, leadRepo, commHistRepo, leadTagRepo, scoringService, hub),
		Notification:        handler.NewNotificationHandler(notificationRepo),
		Deal:                handler.NewDealHandler(dealRepo, invoiceRepo, auditLogRepo),
		Invoice:             handler.NewInvoiceHandler(invoiceRepo, dealRepo, companySettingsRepo),
		Property:            handler.NewPropertyHandler(bos24Client, companySettingsRepo),
		Settings:            handler.NewSettingsHandler(settingsRepo, companySettingsRepo),
		Email:               handler.NewEmailHandler(emailService, emailRepo),
		RentalProperty:      handler.NewRentalPropertyHandler(rentalPropertyRepo),
		Tenant:              handler.NewTenantHandler(tenantRepo),
		LeaseTemplate:       handler.NewLeaseTemplateHandler(leaseTemplateRepo),
		Lease:               handler.NewLeaseHandler(leaseRepo),
		Payment:             handler.NewPaymentHandler(paymentRepo),
		BankIntegration:     handler.NewBankIntegrationHandler(bankIntegrationRepo),
		BankStatement:       handler.NewBankStatementHandler(bankStatementRepo),
		PaymentConfirmation: handler.NewPaymentConfirmationHandler(paymentConfirmationRepo, paymentConfirmationService),
		Analytics:           handler.NewAnalyticsHandler(repo.NewAnalyticsRepository(pool)),
		Expense:             handler.NewExpenseHandler(expenseRepo),
		Inspection:          handler.NewInspectionHandler(inspectionTemplateRepo, inspectionRepo),
		Maintenance:         handler.NewMaintenanceTaskHandler(maintenanceRepo),
		LeaseRenewal:        handler.NewLeaseRenewalHandler(leaseRenewalRepo, renewalTemplateRepo, renewalCommLogRepo),
		Document:           handler.NewDocumentHandler(documentRepo, auditLogRepo),
		ApiKey:              handler.NewApiKeyHandler(apiKeyRepo),
		PublicLead:          handler.NewPublicLeadHandler(contactRepo, leadRepo, dispatcher),
		WebhookSub:          handler.NewWebhookSubHandler(webhookRepo, dispatcher),
		Billing:             handler.NewBillingHandler(billingRepo, companySettingsRepo, stripeCfg),
		MessageTemplate:     handler.NewMessageTemplateHandler(messageTemplateRepo),
		AuditLog:           handler.NewAuditHandler(auditLogRepo),
	}

	// ── Fiber app ────────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "Masaar CRM",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "internal server error"
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				msg = e.Message
			}
			if code == fiber.StatusInternalServerError {
				log.Printf("internal error: %v", err)
			}
			return c.Status(code).JSON(fiber.Map{"error": msg})
		},
	})

	app.Use(helmet.New())
	app.Use(recover.New())
	app.Use(middleware.PIISafeLogger())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PATCH,DELETE,OPTIONS",
		AllowCredentials: cfg.AllowedOrigins != "*",
	}))

	api.RegisterRoutes(app, handlers, hub, cfg, rdb, pool, apiKeyRepo, billingRepo, companyRepo)

	// ── Background Jobs ──────────────────────────────────────────────────────

	go func() {
		ticker := time.NewTicker(12 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			log.Println("Running payment reminder generation job...")
			companies, err := companyRepo.List(ctx)
			if err != nil {
				log.Printf("Error fetching companies: %v", err)
				cancel()
				continue
			}
			for _, company := range companies {
				if err := paymentReminderService.GenerateReminders(ctx, company.ID); err != nil {
					log.Printf("Error generating reminders for company %s: %v", company.ID, err)
				}
			}
			cancel()
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			log.Println("Running payment reminder delivery job...")
			companies, err := companyRepo.List(ctx)
			if err != nil {
				log.Printf("Error fetching companies: %v", err)
				cancel()
				continue
			}
			for _, company := range companies {
				if err := paymentReminderService.SendPendingReminders(ctx, company.ID); err != nil {
					log.Printf("Error sending reminders for company %s: %v", company.ID, err)
				}
			}
			cancel()
		}
	}()

	// Trial expiry checker — runs every 6 hours
	go func() {
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			log.Println("Checking for expired trials...")
			expired, err := companyRepo.ListExpiredTrials(ctx)
			if err != nil {
				log.Printf("Error listing expired trials: %v", err)
				cancel()
				continue
			}
			for _, company := range expired {
				if err := companyRepo.EndTrial(ctx, company.ID); err != nil {
					log.Printf("Error ending trial for company %s: %v", company.ID, err)
				} else {
					log.Printf("Trial ended for company %s (%s)", company.ID, company.Name)
				}
			}
			cancel()
		}
	}()

	// ── Graceful shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Masaar CRM starting on :%s", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")
	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("bye")
}
