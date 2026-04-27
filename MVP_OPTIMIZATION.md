# MVP Resource Optimization & Google Gemini Integration

For MVP with expected max 100 users in 6 months, we should minimize resource usage and use cloud APIs.

---

## 📊 Resource Reallocation

### Current (Heavy) vs MVP (Optimized)

| Service | Heavy | MVP | Savings | Notes |
|---------|-------|-----|---------|-------|
| PostgreSQL | 2 CPU, 8GB | 0.5 CPU, 2GB | **75%** | Sufficient for 100 users |
| Redis | 1 CPU, 2GB | 0.25 CPU, 512MB | **75%** | Cache only, can be small |
| Ollama | 4 CPU, 8GB | REMOVE | **12GB RAM!** | Use Gemini API instead |
| API | 2 CPU, 2GB | 1 CPU, 1GB | **50%** | Light MVP traffic |
| Frontend | 1 CPU, 1GB | 0.5 CPU, 512MB | **50%** | Static with minimal load |

### Resource Summary

**Before (Heavy):**
- Total: 10 CPU, 21GB
- OS + headroom: 2 CPU, 26GB
- **Total used: 12 CPU, 47GB** ✓

**After (MVP):**
- Total: 2.25 CPU, 4.5GB
- OS + headroom: 1 CPU, 10GB
- **Total: ~3.25 CPU, 14.5GB**
- **Freed up: 8.75 CPU, 32.5GB (70% savings)** 🎉

---

## 🚀 Google Gemini Integration

### Step 1: Get Google Gemini API Key

```bash
# 1. Go to https://makersuite.google.com/app/apikey
# 2. Click "Create API Key"
# 3. Copy the key
# 4. Add to .env:
GEMINI_API_KEY=your-key-here
```

### Step 2: Update Environment

Add to `.env`:
```bash
# Google Gemini (replaces local Ollama)
GEMINI_API_KEY=your-api-key-here
GEMINI_MODEL=gemini-1.5-flash  # Faster, cheaper for MVP
# Alternative: gemini-1.5-pro (more capable, higher cost)
```

### Step 3: Create Gemini Service

Create `internal/ai/gemini.go`:

```go
package ai

import (
	"context"
	"fmt"
	"github.com/google/generative-ai-go/client"
	"github.com/google/generative-ai-go/genai"
	"log"
)

type GeminiClient struct {
	client *genai.Client
	model  string
}

func NewGeminiClient(apiKey, model string) (*GeminiClient, error) {
	ctx := context.Background()
	c, err := client.NewClient(ctx, &client.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	return &GeminiClient{
		client: c,
		model:  model,
	}, nil
}

// Summarize thread messages using Gemini
func (gc *GeminiClient) Summarize(ctx context.Context, messages string) (string, error) {
	model := gc.client.GenerativeModel(gc.model)
	model.SetTemperature(0.3) // Lower = more focused

	resp, err := model.GenerateContent(ctx, genai.Text(
		fmt.Sprintf("Summarize this customer conversation concisely (max 2 sentences):\n\n%s", messages),
	))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
	}
	return "", fmt.Errorf("no response from Gemini")
}

// Auto-tag messages using Gemini
func (gc *GeminiClient) ClassifyMessage(ctx context.Context, message string) (string, error) {
	model := gc.client.GenerativeModel(gc.model)
	model.SetTemperature(0)

	resp, err := model.GenerateContent(ctx, genai.Text(
		fmt.Sprintf("Classify this message as: lead-inquiry, complaint, positive, other\nMessage: %s\nRespond with one word only.", message),
	))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
	}
	return "", fmt.Errorf("no response from Gemini")
}

func (gc *GeminiClient) Close() error {
	return gc.client.Close()
}
```

Add to `go.mod`:
```bash
go get github.com/google/generative-ai-go
```

### Step 4: Update Server Initialization

In `cmd/server/main.go`, replace Ollama with Gemini:

```go
// ── AI client ────────────────────────────────────────────────────────────
// Use Google Gemini instead of local Ollama for MVP (cost-effective)
var aiClient interface{}
if cfg.GeminiAPIKey != "" {
	geminiClient, err := ai.NewGeminiClient(cfg.GeminiAPIKey, cfg.GeminiModel)
	if err != nil {
		log.Fatalf("gemini initialization failed: %v", err)
	}
	aiClient = geminiClient
	log.Println("✓ Google Gemini API enabled")
} else {
	log.Println("⚠ AI features disabled - no Gemini API key")
}
```

### Step 5: Update Config

Add to `internal/config/config.go`:

```go
type Config struct {
	// ... existing fields ...
	
	// Google Gemini API (replaces Ollama)
	GeminiAPIKey string
	GeminiModel  string // default: "gemini-1.5-flash"
}

func Load() *Config {
	return &Config{
		// ... existing ...
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		GeminiModel:  getEnv("GEMINI_MODEL", "gemini-1.5-flash"),
	}
}
```

---

## 🐳 MVP Docker Compose (Lightweight)

Create `docker-compose.mvp.yml`:

```yaml
version: "3.9"

services:
  postgres:
    image: postgres:16-alpine
    container_name: masaar-postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: masaar
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: masaar
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U masaar"]
      interval: 10s
      timeout: 5s
      retries: 5
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 2G
        reservations:
          cpus: '0.25'
          memory: 1G
    logging:
      driver: "json-file"
      options:
        max-size: "20m"
        max-file: "3"

  redis:
    image: redis:7-alpine
    container_name: masaar-redis
    restart: unless-stopped
    ports:
      - "6380:6380"
    command: redis-server --requirepass ${REDIS_PASSWORD} --port 6380 --maxmemory 512mb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-p", "6380", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    deploy:
      resources:
        limits:
          cpus: '0.25'
          memory: 512M
        reservations:
          cpus: '0.1'
          memory: 256M
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "2"

  # Ollama REMOVED - using Google Gemini API instead
  # Saves: 4 CPU, 8GB RAM, 5GB disk

  api:
    build:
      context: .
      dockerfile: docker/Dockerfile
    container_name: masaar-api
    restart: unless-stopped
    ports:
      - "8080:8080"
    env_file:
      - .env
    environment:
      DATABASE_URL: postgres://masaar:${POSTGRES_PASSWORD}@postgres:5432/masaar?sslmode=disable
      REDIS_URL: redis://:${REDIS_PASSWORD}@redis:6380
      GEMINI_API_KEY: ${GEMINI_API_KEY}
      GEMINI_MODEL: ${GEMINI_MODEL:-gemini-1.5-flash}
      # No Ollama environment variables needed
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/api/v1/stats"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
        max-file: "5"

  web:
    build:
      context: .
      dockerfile: docker/Dockerfile.web
    container_name: masaar-web
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      NEXT_PUBLIC_API_URL: http://localhost:8080
    depends_on:
      - api
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:3000"]
      interval: 30s
      timeout: 10s
      retries: 2
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M
    logging:
      driver: "json-file"
      options:
        max-size: "20m"
        max-file: "2"

volumes:
  postgres_data:
  redis_data:
```

---

## 📋 MVP vs Production Comparison

| Aspect | MVP | Production |
|--------|-----|-----------|
| **AI Model** | Google Gemini API (cloud) | Local Ollama OR Gemini |
| **Cost** | $0.075/1M input tokens | Free (local) or pay-per-use |
| **CPU** | 2.25 | 10+ |
| **RAM** | 4.5GB | 21GB+ |
| **Startup Time** | Instant | 5-10 min (Ollama model pull) |
| **Scaling** | API calls (unlimited) | Local model (capped) |
| **Maintenance** | None | Model updates |

---

## 🚀 How to Deploy MVP

### Option A: Start Fresh (Recommended for MVP)

```bash
# SSH to server
ssh root@109.123.240.239
cd /opt/masaar-crm

# Use MVP docker-compose
docker compose -f docker-compose.mvp.yml up -d

# Monitor
docker compose -f docker-compose.mvp.yml ps
```

### Option B: Modify Existing Setup

If already deployed with Ollama:

```bash
# Stop Ollama (free up resources)
docker compose down ollama

# Update docker-compose.yml to remove ollama service
nano docker-compose.yml
# Delete the ollama service block

# Restart with Gemini
docker compose up -d api
```

---

## 💰 Cost Analysis (6 months, 100 users)

### Google Gemini Pricing (as of 2026)

**Input tokens:** $0.075 / 1M tokens
**Output tokens:** $0.30 / 1M tokens

### Estimated Usage (MVP)

Assuming:
- 100 users max
- 10 messages/day per user average
- Summary: 300 tokens input, 50 output
- Auto-tag: 100 tokens input, 20 output

**Daily:**
- Summaries: 100 users × 1 summary = 100 × (300+50) = 35,000 tokens
- Auto-tag: 100 users × 10 messages = 1,000 × (100+20) = 120,000 tokens
- **Total daily: ~155,000 tokens**

**Monthly:** 155,000 × 30 = 4,650,000 tokens
**Cost per month:** ~$0.40 (4.65M × $0.075/M)
**6 months:** ~$2.40

**Gemini cost for 6 months:** ~$3 (including 20% buffer)

vs.

**Ollama cost for 6 months:**
- 8GB RAM saved: ~$10/month = $60
- 4 CPU cores saved: ~$15/month = $90
- **Ollama local: FREE but uses $150 worth of compute**

**Recommendation:** Use Gemini for MVP, save resources, pay $3. Win-win! 🎉

---

## 🔧 Implementation Checklist

- [ ] Get Google Gemini API key from makersuite.google.com
- [ ] Add `GEMINI_API_KEY` to .env
- [ ] Update `internal/config/config.go` with Gemini settings
- [ ] Create `internal/ai/gemini.go` with Gemini client
- [ ] Update `cmd/server/main.go` to use Gemini instead of Ollama
- [ ] Run `go mod tidy && go mod download`
- [ ] Update docker-compose to use MVP version
- [ ] Remove Ollama from docker-compose
- [ ] Test AI features (summarize, auto-tag)
- [ ] Deploy to production

---

## 🎯 Benefits of Gemini for MVP

✅ **Resource Efficient**
- Frees up 8GB RAM, 4 CPU cores
- Runs smoothly on shared server
- No startup delays

✅ **Cost Effective**
- Pay only for actual usage (~$3 for 6 months)
- No local model management
- No disk space for models

✅ **Scalable**
- API handles unlimited concurrent requests
- No performance degradation as users grow
- Auto-scales with traffic

✅ **Maintenance Free**
- Google handles updates
- No model management
- Always latest model version

✅ **Better for Small Teams**
- No ML infrastructure needed
- Focus on product features
- Easy to switch models (flash vs pro)

---

## 📈 Migration Path (When You Grow)

### Month 1-3 (MVP: Use Gemini)
- Validate product with users
- Measure AI feature usage
- Optimize prompts

### Month 4-6 (Scale: Could switch)
- If budget allows: Run Ollama locally
- If growth justified: Upgrade to Gemini Pro
- If privacy critical: Self-host Ollama

### Month 7+ (Production)
- Assess user demand for AI
- Choose based on: cost vs latency vs privacy
- Consider hybrid approach

---

## ⚠️ Important Notes

### Gemini API Key Security
```bash
# NEVER commit API key to git
echo "GEMINI_API_KEY=your-key" >> .env
git add .env.example  # Only example
git ignore .env       # Never commit actual keys

# On server, keep .env secure
chmod 600 /opt/masaar-crm/.env
```

### Rate Limits
Google Gemini free tier:
- 60 requests per minute (plenty for MVP)
- Scales to thousands with Pro plan

### Fallback
If Gemini is down (rare):
```go
if cfg.GeminiAPIKey == "" {
    return "AI features unavailable"
}
```

---

## 📝 Updated .env.example

```bash
# ... existing variables ...

# AI Service - Choose one:
# Option 1: Google Gemini (Recommended for MVP)
GEMINI_API_KEY=your-api-key-here
GEMINI_MODEL=gemini-1.5-flash

# Option 2: Local Ollama (Removed for MVP)
# OLLAMA_BASE_URL=http://ollama:11434
# OLLAMA_MODEL=llama3
```

---

## 🚀 Summary

| Factor | Decision | Reason |
|--------|----------|--------|
| **AI Model** | Google Gemini | Saves 8GB RAM, costs $3/6mo |
| **Ollama** | Remove for MVP | Not needed, use API instead |
| **Resources** | MVP config | 2.25 CPU, 4.5GB RAM |
| **Scaling** | API handles it | Unlimited concurrent requests |
| **Cost** | ~$3 for 6 months | Near-free MVP |
| **Path** | Easy to change | Can add Ollama later anytime |

---

**This approach is perfect for your MVP: minimal resources, maximum efficiency, near-zero AI cost.** ✅
