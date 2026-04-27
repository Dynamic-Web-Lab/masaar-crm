# Masaar CRM — Production Deployment Plan

**Target Server:** vmi3243176.contaboserver.com (109.123.240.239)
**OS:** Ubuntu 24.04.4 LTS
**Status:** ✅ Docker-ready (v29.4.0, Compose v5.1.3)

---

## Server Capacity Analysis

| Resource | Total | Used | Available | Status |
|----------|-------|------|-----------|--------|
| **CPU** | 12 cores | N/A | 12 cores | ✅ Excellent |
| **RAM** | 47 GB | 1.3 GB | 45 GB | ✅ Excellent |
| **Disk** | 242 GB | 168 GB | 75 GB | ⚠️ Monitor (70% full) |
| **Network** | 1 Gbps | 2 active | Available | ✅ Good |

### Disk Space Advisory
- Current usage: 70% (consider cleanup of old Docker images/volumes)
- Masaar requirements: ~15-20GB (PostgreSQL, Redis, Ollama models, app)
- Recommendation: Clean up unused Docker artifacts before deployment

---

## Port Allocation Strategy

**Existing Services (DO NOT TOUCH):**
```
:80, :443   → nginx (BuyOrSell24 reverse proxy)
:8000       → buyorsell24-api (Python)
:8001       → buyorsell24-go-api (Go)
:6379       → buyorsell24-redis
:7700       → meilisearch
```

**Masaar CRM Services (NEW):**
```
:8080       → Masaar API (Go/Fiber)
:3000       → Masaar Frontend (Next.js)
:5432       → Masaar PostgreSQL (internal only)
:6380       → Masaar Redis (internal only)
:11434      → Masaar Ollama (internal only)
```

**Routing via Nginx:**
```
https://api.yourdomain.com          → localhost:8080 (Masaar API)
https://crm.yourdomain.com          → localhost:3000 (Masaar Frontend)
https://old-domain.com/api          → localhost:8000 (BuyOrSell24 API)
```

---

## Deployment Steps

### Step 1: Prepare Server Environment

```bash
# SSH into server (you're already there)

# 1a. Clean up Docker space (optional but recommended)
docker image prune -a --force     # Remove unused images
docker volume prune --force       # Remove unused volumes
docker system prune -a --force    # Full cleanup

# 1b. Create Masaar deployment directory
mkdir -p /opt/masaar-crm
cd /opt/masaar-crm

# 1c. Create directories for volumes
mkdir -p volumes/{postgres,redis,ollama}
chmod 755 volumes/*
```

### Step 2: Clone Repository

```bash
# Clone Masaar CRM repo
git clone https://github.com/maidulcu/masaar-crm.git .
# OR if you have a private repo:
git clone git@github.com:yourusername/masaar-crm.git .

cd /opt/masaar-crm
```

### Step 3: Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit with production values
nano .env
```

**Production .env values:**
```bash
# Server
PORT=8080
APP_ENV=production

# Database (use local container)
DATABASE_URL=postgres://masaar:SECURE_PASSWORD_HERE@postgres:5432/masaar?sslmode=disable

# Redis (use local container)
REDIS_URL=redis://redis:6380

# JWT (CRITICAL: Change this to random 32+ char string)
JWT_SECRET=generate-with-openssl-rand-base64-32
JWT_ACCESS_EXPIRY_MIN=15
JWT_REFRESH_EXPIRY_DAYS=7

# WhatsApp Business API
WA_VERIFY_TOKEN=your-webhook-token
WA_API_VERSION=v19.0
WA_PHONE_NUMBER_ID=your-phone-id
WA_ACCESS_TOKEN=your-access-token
WA_BASE_URL=https://graph.instagram.com

# Ollama (local LLM)
OLLAMA_BASE_URL=http://ollama:11434
OLLAMA_MODEL=llama3

# SMTP (optional - for invoices, proposals)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_FROM_NAME=Masaar CRM

# Optional: Real estate API
BOS24_API_TOKEN=
```

### Step 4: Update Docker Compose for Production

The default `docker-compose.yml` is dev-focused. Create a production override:

```bash
# Copy the template
cp docker/docker-compose.prod.yml docker-compose.prod.yml
```

**Edit `/opt/masaar-crm/docker-compose.yml` to use custom ports:**

```yaml
version: "3.9"

services:
  postgres:
    image: postgres:16
    container_name: masaar-postgres
    restart: always
    environment:
      POSTGRES_USER: masaar
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}  # from .env
      POSTGRES_DB: masaar
    ports:
      - "5432:5432"  # Internal only (no bind to 0.0.0.0)
    volumes:
      - ./volumes/postgres:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U masaar"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: masaar-redis
    restart: always
    command: redis-server --requirepass ${REDIS_PASSWORD} --port 6380
    ports:
      - "6380:6380"  # Internal only
    volumes:
      - ./volumes/redis:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-p", "6380", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  ollama:
    image: ollama/ollama:latest
    container_name: masaar-ollama
    restart: always
    ports:
      - "11434:11434"  # Internal only
    volumes:
      - ./volumes/ollama:/root/.ollama
    entrypoint: ["/bin/sh", "-c", "ollama serve & sleep 10 && ollama pull ${OLLAMA_MODEL:-llama3} && wait"]

  api:
    build:
      context: .
      dockerfile: docker/Dockerfile
    container_name: masaar-api
    restart: always
    environment:
      PORT: 8080
      APP_ENV: production
      DATABASE_URL: ${DATABASE_URL}
      REDIS_URL: ${REDIS_URL}
      JWT_SECRET: ${JWT_SECRET}
      WA_VERIFY_TOKEN: ${WA_VERIFY_TOKEN}
      WA_ACCESS_TOKEN: ${WA_ACCESS_TOKEN}
      WA_PHONE_NUMBER_ID: ${WA_PHONE_NUMBER_ID}
      WA_BASE_URL: ${WA_BASE_URL}
      OLLAMA_BASE_URL: http://ollama:11434
      OLLAMA_MODEL: ${OLLAMA_MODEL}
      SMTP_HOST: ${SMTP_HOST}
      SMTP_PORT: ${SMTP_PORT}
      SMTP_USER: ${SMTP_USER}
      SMTP_PASSWORD: ${SMTP_PASSWORD}
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    logging:
      driver: "json-file"
      options:
        max-size: "20m"
        max-file: "5"

  web:
    build:
      context: .
      dockerfile: docker/Dockerfile.web
    container_name: masaar-web
    restart: always
    environment:
      NEXT_PUBLIC_API_URL: https://api.yourdomain.com  # Change this
    ports:
      - "3000:3000"
    depends_on:
      - api
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

volumes:
  postgres_data:
  redis_data:
  ollama_data:
```

### Step 5: Start Masaar CRM

```bash
cd /opt/masaar-crm

# Build and start (first run takes ~10-15 min for Ollama model pull)
docker compose up -d

# Watch logs
docker compose logs -f api

# Check status
docker compose ps
```

**Expected output:**
```
NAME              STATUS                 PORTS
masaar-postgres   Up (healthy)           5432/tcp
masaar-redis      Up (healthy)           6380/tcp
masaar-ollama     Up (downloading)       11434/tcp
masaar-api        Up (healthy)           8080/tcp
masaar-web        Up                     3000/tcp
```

### Step 6: Configure Nginx Reverse Proxy

Update your existing nginx config to route Masaar traffic:

```bash
# Edit nginx config
sudo nano /etc/nginx/sites-available/default
# OR if using Docker nginx:
docker exec buyorsell24-nginx /bin/sh -c "cat /etc/nginx/conf.d/default.conf"
```

**Add these upstream blocks and server blocks:**

```nginx
# Upstream Masaar services
upstream masaar_api {
    server localhost:8080;
}

upstream masaar_web {
    server localhost:3000;
}

# Masaar API subdomain
server {
    listen 80;
    listen [::]:80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://masaar_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Masaar CRM Frontend subdomain
server {
    listen 80;
    listen [::]:80;
    server_name crm.yourdomain.com;

    location / {
        proxy_pass http://masaar_web;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

Then:
```bash
# Test config
sudo nginx -t

# Reload
sudo systemctl reload nginx
# OR if Docker:
docker exec buyorsell24-nginx nginx -s reload
```

### Step 7: SSL/TLS (Let's Encrypt)

```bash
# Install certbot
sudo apt-get install certbot python3-certbot-nginx

# Get certificates
sudo certbot certonly --nginx -d api.yourdomain.com -d crm.yourdomain.com

# Auto-renewal (already enabled in Ubuntu)
sudo systemctl status certbot.timer
```

Update nginx to use SSL:

```nginx
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;
    
    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;
    
    # ... rest of config
}
```

---

## Post-Deployment Verification

### 1. Health Checks

```bash
# API health
curl http://localhost:8080/api/v1/stats

# Frontend
curl http://localhost:3000

# Database
docker exec masaar-postgres psql -U masaar -d masaar -c "SELECT version();"

# Redis
docker exec masaar-redis redis-cli -p 6380 PING

# Ollama
curl http://localhost:11434/api/tags
```

### 2. Default Login

Once frontend loads:
- **Email:** admin@masaar.local
- **Password:** changeme

⚠️ **CHANGE THIS IMMEDIATELY** in the database or via API.

### 3. Check Logs

```bash
# API logs
docker compose logs api --tail 50

# Frontend logs
docker compose logs web --tail 50

# Database logs
docker compose logs postgres --tail 50
```

---

## Monitoring & Maintenance

### Daily Checks

```bash
# Container status
docker compose ps

# Disk space
df -h /

# Memory usage
free -h

# Docker logs (errors)
docker compose logs --tail 100 | grep -i error
```

### Weekly Tasks

```bash
# Check for updates
docker compose pull

# Prune unused data
docker system prune
```

### Monthly Tasks

```bash
# Backup database
docker exec masaar-postgres pg_dump -U masaar masaar > /opt/masaar-crm/backups/masaar-$(date +%Y%m%d).sql

# Backup .env
cp /opt/masaar-crm/.env /opt/masaar-crm/backups/.env-$(date +%Y%m%d)
```

---

## Troubleshooting

### Ollama taking too long to download

```bash
# Check progress
docker compose logs ollama --tail 20

# If stuck, restart
docker compose restart ollama

# Manual pull (if needed)
docker exec masaar-ollama ollama pull llama3
```

### Database connection errors

```bash
# Check DB is healthy
docker compose ps postgres

# Connect and check
docker exec masaar-postgres psql -U masaar -d masaar -c "\dt"
```

### API not responding

```bash
# Restart API
docker compose restart api

# Check logs
docker compose logs api --tail 50

# Test endpoint
curl -v http://localhost:8080/api/v1/stats
```

### Port conflicts

If ports are already in use:
```bash
# Find what's using port 8080
sudo lsof -i :8080

# Or with netstat
netstat -tulpn | grep 8080
```

---

## DNS Configuration

Once server is ready, update your DNS:

```dns
A Record: api.yourdomain.com         → 109.123.240.239
A Record: crm.yourdomain.com         → 109.123.240.239
```

---

## Backup & Recovery Procedures

### Automated Daily Backups

```bash
# Create backup script
cat > /opt/masaar-crm/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/masaar-crm/backups"
mkdir -p $BACKUP_DIR

# Database backup
docker exec masaar-postgres pg_dump -U masaar masaar | gzip > $BACKUP_DIR/db-$(date +%Y%m%d-%H%M%S).sql.gz

# Keep only last 7 days
find $BACKUP_DIR -name "db-*.sql.gz" -mtime +7 -delete

echo "Backup completed at $(date)" >> $BACKUP_DIR/backup.log
EOF

chmod +x /opt/masaar-crm/backup.sh

# Schedule with cron
crontab -e
# Add: 0 2 * * * /opt/masaar-crm/backup.sh
```

### Full System Restore

```bash
# Stop containers
docker compose down

# Restore database
gunzip < backups/db-YYYYMMDD-HHMMSS.sql.gz | docker exec -i masaar-postgres psql -U masaar masaar

# Start containers
docker compose up -d
```

---

## Security Checklist

- [ ] Change `JWT_SECRET` to random 32+ character string
- [ ] Change `POSTGRES_PASSWORD` in .env
- [ ] Change `REDIS_PASSWORD` in .env
- [ ] Change default admin password (admin@masaar.local / changeme)
- [ ] Enable SSL/TLS (Let's Encrypt configured)
- [ ] Set up firewall rules (only 80, 443 exposed)
- [ ] Configure WhatsApp webhook token
- [ ] Set up automated backups
- [ ] Enable Docker log rotation
- [ ] Monitor disk space (currently 70% full)

---

## Reference

- **API Documentation:** http://localhost:8080/docs (internal only)
- **Swagger Spec:** http://localhost:8080/swagger.json
- **Frontend:** http://localhost:3000 (internal) or https://crm.yourdomain.com
- **Default DB:** masaar (user: masaar)

---

**Deployment completed:** [date]
**Last updated:** 2026-04-27
