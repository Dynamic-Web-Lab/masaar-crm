package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret            string
	JWTAccessExpiryMin   int
	JWTRefreshExpiryDays int

	// WhatsApp (Inbound + Outbound)
	WAVerifyToken   string
	WAAPIVersion    string
	WAPhoneNumberID string
	WAAccessToken   string
	WABaseURL       string
	WAAppSecret     string // used to validate X-Hub-Signature-256 on inbound webhooks

	// AI provider — "ollama" (local) or "gemini" (Google Cloud)
	AIProvider    string
	OllamaBaseURL string
	OllamaModel   string
	GeminiAPIKey  string
	GeminiModel   string

	// BuyOrSell24 (Real Estate API)
	BOS24Token   string
	BOS24BaseURL string

	// Email provider: "smtp" (default) or "azure" (Azure Communication Services)
	EmailProvider  string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFromEmail string
	SMTPFromName string
	// Azure Communication Services (REST API, not SMTP)
	AzureCommEndpoint   string
	AzureCommKey        string
	AzureCommFromAddress string

	// Magic Link (Passwordless Login)
	MagicLinkBaseURL  string
	MagicLinkExpiryMin int
	MagicLinkRateLimit int

	// App
	AppEnv string
	AppURL string // base URL for generating links in emails, e.g. https://crm.yourcompany.ae

	// Multi-tenancy (single-tenant: fixed company UUID for this installation)
	CompanyID string

	// CORS — comma-separated allowed origins; defaults to * in development only
	AllowedOrigins string

	// Stripe billing (optional — leave empty for self-hosted/community)
	StripeSecretKey       string
	StripeWebhookSecret   string
	StripePriceIDStarter  string
	StripePriceIDPro      string
	StripePriceIDBusiness string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	return &Config{
		Port:                 getEnv("PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://masaar:masaar@localhost:5432/masaar?sslmode=disable"),
		RedisURL:             getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:            getEnv("JWT_SECRET", "change-me-in-production"),
		JWTAccessExpiryMin:   getEnvInt("JWT_ACCESS_EXPIRY_MIN", 15),
		JWTRefreshExpiryDays: getEnvInt("JWT_REFRESH_EXPIRY_DAYS", 7),
		WAVerifyToken:        getEnv("WA_VERIFY_TOKEN", "masaar-webhook-token"),
		WAAPIVersion:         getEnv("WA_API_VERSION", "v19.0"),
		WAPhoneNumberID:      getEnv("WA_PHONE_NUMBER_ID", ""),
		WAAccessToken:        getEnv("WA_ACCESS_TOKEN", ""),
		WABaseURL:            getEnv("WA_BASE_URL", "https://graph.instagram.com"),
		WAAppSecret:          getEnv("WA_APP_SECRET", ""),
		AIProvider:           getEnv("AI_PROVIDER", "ollama"),
		OllamaBaseURL:        getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaModel:          getEnv("OLLAMA_MODEL", "llama3"),
		GeminiAPIKey:         getEnv("GEMINI_API_KEY", ""),
		GeminiModel:          getEnv("GEMINI_MODEL", "gemini-2.0-flash"),
		BOS24Token:            getEnv("BOS24_API_TOKEN", ""),
		BOS24BaseURL:          getEnv("BOS24_BASE_URL", "https://data.buyorsell24.com"),
		EmailProvider:         getEnv("EMAIL_PROVIDER", "smtp"),
		SMTPHost:             getEnv("SMTP_HOST", ""),
		SMTPPort:             getEnv("SMTP_PORT", "587"),
		SMTPUser:             getEnv("SMTP_USER", ""),
		SMTPPassword:         getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:        getEnv("SMTP_FROM_EMAIL", "noreply@masaar.local"),
		SMTPFromName:         getEnv("SMTP_FROM_NAME", "Masaar CRM"),
		AzureCommEndpoint:    getEnv("AZURE_COMM_ENDPOINT", ""),
		AzureCommKey:         getEnv("AZURE_COMM_KEY", ""),
		AzureCommFromAddress: getEnv("AZURE_COMM_FROM_ADDRESS", ""),
		MagicLinkBaseURL:     getEnv("MAGIC_LINK_BASE_URL", "http://localhost:3000"),
		MagicLinkExpiryMin:   getEnvInt("MAGIC_LINK_EXPIRY_MIN", 15),
		MagicLinkRateLimit:   getEnvInt("MAGIC_LINK_RATE_LIMIT", 3),
		AppEnv:               getEnv("APP_ENV", "development"),
		AppURL:               getEnv("APP_URL", "http://localhost:3000"),
		CompanyID:             getEnv("APP_COMPANY_ID", "00000000-0000-0000-0000-000000000001"),
		AllowedOrigins:        getEnv("ALLOWED_ORIGINS", "*"),
		StripeSecretKey:       getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:   getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripePriceIDStarter:  getEnv("STRIPE_PRICE_STARTER", ""),
		StripePriceIDPro:      getEnv("STRIPE_PRICE_PRO", ""),
		StripePriceIDBusiness: getEnv("STRIPE_PRICE_BUSINESS", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
