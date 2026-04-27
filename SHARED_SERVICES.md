# Shared Services Architecture (Advanced Optimization)

Since your server already has Redis and Meilisearch running for BuyOrSell24, Masaar can share them!

---

## 🎯 Current Server Services

```
BuyOrSell24:
  - nginx (ports 80, 443)
  - buyorsell24-api (port 8000)
  - buyorsell24-go-api (port 8001)
  - redis:7 (port 6379) ← Can be SHARED
  - meilisearch (port 7700) ← Can be SHARED
```

---

## 🔄 Shared Services Setup

### Option 1: Use Shared Redis (Recommended)

Instead of running Redis in docker-compose, connect to existing:

**In `docker-compose.mvp.yml`, REMOVE redis service:**

```yaml
# REMOVE THIS SECTION:
# redis:
#   image: redis:7-alpine
#   ...
```

**In `.env`, CHANGE Redis URL:**

```bash
# OLD (separate Redis):
REDIS_URL=redis://redis:6380

# NEW (shared Redis on port 6379):
REDIS_URL=redis://redis:6379

# If Redis requires password (check with BuyOrSell24 config):
REDIS_URL=redis://:PASSWORD@redis:6379
```

**Docker network setup:**
```yaml
# In docker-compose.mvp.yml
services:
  api:
    networks:
      - buyorsell24_default  # Join existing network
    
  web:
    networks:
      - buyorsell24_default

networks:
  buyorsell24_default:
    external: true  # Use existing network
```

**Benefits of shared Redis:**
✅ No duplicate Redis running
✅ Saves 0.25 CPU, 512MB RAM
✅ Easier management (one Redis instance)
✅ Shared cache across both apps (good for coordination)

---

### Option 2: Use Shared Meilisearch (Optional)

If you want full-text search for Masaar contacts/leads:

**Add to models for search capability:**

```go
// In internal/repo/contact.go
func (cr *ContactRepository) Search(ctx context.Context, query string) ([]Contact, error) {
    // Use Meilisearch instead of database LIKE queries
    // Much faster for large datasets
}
```

**Docker configuration:**
```yaml
# docker-compose.mvp.yml
api:
  environment:
    MEILISEARCH_URL: http://meilisearch:7700
  networks:
    - buyorsell24_default
```

**Benefits:**
✅ Fast search: 100ms vs 500ms+ database LIKE
✅ Typo tolerance: "contct" finds "contact"
✅ Faceted search: filter by company, status, etc.
✅ Shared instance (no duplication)

---

## 💾 Optimized Shared Architecture

### Without Sharing

```
PostgreSQL   Redis (6379)  Meilisearch
    ↓            ↓              ↓
BuyOrSell24  BuyOrSell24   BuyOrSell24
    ↓            ↓
 Masaar       Redis (6380)   [No search]
```

Resources used: 
- 2 Redis instances
- No search for Masaar

### With Sharing

```
PostgreSQL   Redis (6379)  Meilisearch
    ↓            ↑              ↑
BuyOrSell24  Shared Instances
    ↓            ↑              ↑
 Masaar ─────────┴──────────────┘
```

Resources saved:
- ✅ 1 Redis instance (saves 512MB RAM, 0.25 CPU)
- ✅ Shared search (centralized, efficient)
- ✅ Coordinated caching (same instance)

---

## 🚀 Ultra-Lightweight Docker Compose (With Sharing)

Create `docker-compose.mvp-shared.yml`:

```yaml
version: "3.9"

services:
  # PostgreSQL only (not shared - Masaar needs its own database)
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

  # Redis and Meilisearch: SHARED from BuyOrSell24 (not defined here)
  
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
      # Connect to shared Redis (same as BuyOrSell24)
      REDIS_URL: redis://redis:6379
      # Connect to shared Meilisearch
      MEILISEARCH_URL: http://meilisearch:7700
      GEMINI_API_KEY: ${GEMINI_API_KEY}
      GEMINI_MODEL: ${GEMINI_MODEL:-gemini-1.5-flash}
    networks:
      - buyorsell24_default
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/api/v1/stats"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M

  web:
    build:
      context: .
      dockerfile: docker/Dockerfile.web
    container_name: masaar-web
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL}
      NODE_ENV: production
    networks:
      - buyorsell24_default
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

volumes:
  postgres_data:

networks:
  # Use existing BuyOrSell24 network for service communication
  buyorsell24_default:
    external: true
```

---

## 📊 Resource Comparison (All Scenarios)

| Scenario | CPU | RAM | Cost | Notes |
|----------|-----|-----|------|-------|
| **Standard (Heavy)** | 10 | 21GB | High | Full separate stack |
| **MVP (Light)** | 2.25 | 4.5GB | Low | Gemini + own Redis |
| **MVP (Shared)** | 1.75 | 2.5GB | Very Low | Gemini + shared services |
| **Server Headroom** | 10.25 | 44.5GB | - | For BuyOrSell24 + others |

**Shared MVP Savings:**
- ✅ 70% less CPU than standard
- ✅ 88% less RAM than standard  
- ✅ Minimal footprint on shared server
- ✅ Easy to scale later

---

## ✅ How to Setup Shared Services

### Step 1: Check Existing Services

```bash
ssh root@109.123.240.239

# Check Redis status
docker ps | grep redis
# Should show: buyorsell24-redis (port 6379)

# Check Meilisearch status
docker ps | grep meilisearch
# Should show: buyorsell24-meilisearch (port 7700)

# Verify they're running
docker exec buyorsell24-redis redis-cli PING
docker exec buyorsell24-meilisearch curl -s http://localhost:7700/health
```

### Step 2: Update Masaar Configuration

```bash
cd /opt/masaar-crm

# Use shared services compose
docker compose -f docker-compose.mvp-shared.yml down
docker compose -f docker-compose.mvp-shared.yml up -d

# Update .env
nano .env
# Change: REDIS_URL=redis://redis:6379
# (From 6380 to 6379 - shared instance)
```

### Step 3: Test Connectivity

```bash
# Check API can reach shared Redis
docker exec masaar-api redis-cli -h redis -p 6379 PING

# Check API can reach shared Meilisearch
docker exec masaar-api curl -s http://meilisearch:7700/health
```

---

## 🔒 Important Considerations

### Data Isolation
```
GOOD ✅
- Each app has own database (PostgreSQL)
- Separate auth, separate users
- Safe to share cache layer (Redis)

BAD ❌
- Sharing database directly
- Sharing authentication tokens
- Cross-app access without isolation
```

### Redis Key Prefix

To avoid conflicts in shared Redis:

```go
// In internal/repo - use namespaced keys
const RedisKeyPrefix = "masaar:"

func (ur *UserRepository) CacheKey(userID string) string {
    return fmt.Sprintf("%suser:%s", RedisKeyPrefix, userID)
}
```

### Cache Invalidation

If both apps use Redis:

```bash
# BuyOrSell24 keys: "bos24:*"
# Masaar keys: "masaar:*"

# List Masaar keys only
redis-cli --pattern "masaar:*"

# Flush only Masaar keys (careful!)
redis-cli --scan --pattern "masaar:*" | xargs redis-cli DEL
```

---

## 🎯 Recommendation for MVP

**Use `docker-compose.mvp-shared.yml`:**

1. **Minimal resources:** 1.75 CPU, 2.5GB RAM
2. **Shared infrastructure:** Leverages existing services
3. **Easy to scale:** Add dedicated Redis/Meilisearch later
4. **Cost effective:** ~$3/month for Gemini API
5. **Safe:** Each app has own database

---

## 📈 Future Scaling Path

### MVP (Now)
```
Shared Redis (6379) + Shared Meilisearch
├─ BuyOrSell24
└─ Masaar
```

### Early Growth (Month 3-6)
```
Shared Redis (6379) + Shared Meilisearch
├─ BuyOrSell24
├─ Masaar
└─ New Project
```

### Scale (Month 6+)
```
Redis (6379) - BuyOrSell24
Redis (6380) - Masaar
Redis (6381) - New Project
Meilisearch - Shared

Or: Dedicated clusters if high demand
```

---

## ✨ Summary

| Feature | Standard | MVP Light | MVP Shared |
|---------|----------|-----------|-----------|
| Resources | 10 CPU, 21GB | 2.25 CPU, 4.5GB | **1.75 CPU, 2.5GB** |
| Own Redis | Yes | Yes | No (shared) |
| Own Database | Yes | Yes | Yes |
| Gemini API | No | Yes | Yes |
| Search | Database LIKE | No search | Meilisearch |
| Cost | High | ~$3 | **~$3 + shared** |

**Recommendation:** Use `docker-compose.mvp-shared.yml` for maximum efficiency! 🚀
