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
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/maidulcu/masaar-crm/internal/ai"
	"github.com/maidulcu/masaar-crm/internal/api"
	"github.com/maidulcu/masaar-crm/internal/api/handler"
	"github.com/maidulcu/masaar-crm/internal/bos24"
	"github.com/maidulcu/masaar-crm/internal/config"
	"github.com/maidulcu/masaar-crm/internal/email"
	"github.com/maidulcu/masaar-crm/internal/repo"
	"github.com/maidulcu/masaar-crm/internal/whatsapp"
	"github.com/maidulcu/masaar-crm/internal/ws"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

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

	// ── Email service (optional SMTP integration) ────────────────────────────
	emailService := email.NewService(&email.Config{
		SMTPHost:     cfg.SMTPHost,
		SMTPPort:     cfg.SMTPPort,
		SMTPUser:     cfg.SMTPUser,
		SMTPPassword: cfg.SMTPPassword,
		FromEmail:    cfg.SMTPFromEmail,
		FromName:     cfg.SMTPFromName,
	})

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

	// ── AI client ────────────────────────────────────────────────────────────
	ollamaClient := ai.NewClient(cfg.OllamaBaseURL, cfg.OllamaModel)

	// ── Scoring service for automatic lead scoring ───────────────────────────
	scoringService := ai.NewScoringService(leadRepo, commHistRepo, leadTagRepo)

	// ── Tagging service for auto-tagging on messages ────────────────────────
	taggingService := ai.NewTaggingService(ollamaClient, leadRepo, waRepo, leadTagRepo, contactRepo)

	// ── BuyOrSell24 client (optional real estate integration) ─────────────────
	var bos24Client *bos24.Client
	// Try to load token from database first, fall back to .env
	dbToken, err := settingsRepo.GetBOS24Token(context.Background())
	if err != nil {
		log.Println("no BOS24 token in database, checking .env")
		dbToken = cfg.BOS24Token
	}
	if bos24.IsEnabled(dbToken) {
		bos24Client = bos24.NewClient(dbToken, rdb)
		log.Println("BuyOrSell24 integration enabled")
	}

	// ── Handlers ─────────────────────────────────────────────────────────────
	handlers := &api.Handlers{
		Auth:         handler.NewAuthHandler(userRepo, rdb, cfg),
		User:         handler.NewUserHandler(userRepo),
		Stats:        handler.NewStatsHandler(statsRepo),
		Contact:      handler.NewContactHandler(contactRepo),
		Lead:             handler.NewLeadHandler(leadRepo, contactRepo, commHistRepo, scoringService, hub),
		WhatsApp:         handler.NewWhatsAppHandler(waRepo, contactRepo, taggingService, hub, cfg),
		WhatsAppOutbound: handler.NewWhatsAppOutboundHandler(whatsappSender, outboundRepo, waRepo),
		AI:               handler.NewAIHandler(ollamaClient, contactRepo, leadRepo, waRepo),
		Message:          handler.NewMessageHandler(ollamaClient, waRepo, contactRepo, leadRepo, commHistRepo, leadTagRepo, scoringService, hub),
		Notification: handler.NewNotificationHandler(notificationRepo),
		Deal:         handler.NewDealHandler(dealRepo, invoiceRepo),
		Invoice:      handler.NewInvoiceHandler(invoiceRepo, dealRepo, companySettingsRepo),
		Property:     handler.NewPropertyHandler(bos24Client),
		Settings:     handler.NewSettingsHandler(settingsRepo, companySettingsRepo),
		Email:        handler.NewEmailHandler(emailService, emailRepo),
	}

	// ── Fiber app ────────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "Masaar CRM",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(helmet.New())
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} ${method} ${path} ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS",
	}))

	api.RegisterRoutes(app, handlers, hub, cfg, rdb)

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
