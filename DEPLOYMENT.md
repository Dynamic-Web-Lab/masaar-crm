# Masaar CRM — Production Deployment Guide

**Live URL:** https://masaar.dynamicweblab.com  
**Server:** 109.123.240.239 (Contabo VPS, Ubuntu 24.04)  
<<<<<<< HEAD
**Project dir:** `/root/projects/masaar-crm`  
**SSH key:** `/Users/maidul/info/contabo/backup-buyorsell24-key`  
**Compose file:** `docker-compose.server.yml`

---

## ⚠️ Critical — Shared Server Rules

This server hosts **4 separate projects**. Masaar shares the nginx container and Redis instance with BuyOrSell24. **Never run `docker compose down` without specifying containers** — it will take down other projects.

### Always use targeted commands:
```bash
# ✅ CORRECT — only restarts Masaar containers
docker compose -f docker-compose.server.yml up -d api web

# ❌ WRONG — will stop ALL containers on server including buyorsell24
docker compose down
docker stop $(docker ps -q)
=======
**Project dir on server:** `/root/projects/masaar-crm`  
**SSH key:** `/Users/maidul/info/contabo/backup-buyorsell24-key`  
**Docker Compose file:** `docker-compose.server.yml`

---

## Quick Reference — Container Names

| Container | Role | Port |
|---|---|---|
| `masaar-api` | Go/Fiber backend | 8080 (internal) |
| `masaar-web` | Next.js frontend | 3000 (internal) |
| `masaar-postgres` | PostgreSQL 16 | 5432 (internal) |
| `buyorsell24-redis` | Shared Redis (DB 1 = Masaar) | 6379 (internal) |
| `ollama` | Local LLM for PII operations | 11434 (internal) |

---

## Method A — Deploy Code Changes (Git Pull + Rebuild)

Use this when you've merged code changes into `main` on GitHub.

### 1. SSH into the server

```bash
ssh -i /Users/maidul/info/contabo/backup-buyorsell24-key root@109.123.240.239
```

### 2. Pull latest code

```bash
cd /root/projects/masaar-crm
git pull origin main
```

### 3. Rebuild changed containers (run in background — survives disconnects)

**API only (Go backend changed):**
```bash
nohup docker compose -f docker-compose.server.yml build api > /tmp/masaar-build.log 2>&1 &
# Watch progress:
tail -f /tmp/masaar-build.log
```

**Web only (Next.js frontend changed):**
```bash
nohup docker compose -f docker-compose.server.yml build web > /tmp/masaar-web-build.log 2>&1 &
tail -f /tmp/masaar-web-build.log
```

**Both (full rebuild):**
```bash
nohup docker compose -f docker-compose.server.yml build api web > /tmp/masaar-build.log 2>&1 &
tail -f /tmp/masaar-build.log
```

### 4. Restart containers

**API only (Go code changed):**
```bash
cd /root/projects/masaar-crm
docker compose -f docker-compose.server.yml up -d api web
```

### 5. Verify

```bash
docker ps --format 'table {{.Names}}\t{{.Status}}'
docker logs masaar-api --tail 20
```

Expected API log lines:
```
AI cloud (non-PII): Gemini gemini-2.0-flash
SMSCountry integration enabled
Masaar CRM starting on :8080
>>>>>>> b78c68cc07c568270c9d85ed9c6c39668b24debd
```

---

<<<<<<< HEAD
## All Containers on This Server

| Container | Project | Network | Public? |
|---|---|---|---|
| `buyorsell24-nginx` | BuyOrSell24 | multiple | ✅ :80/:443 — **the shared reverse proxy** |
| `buyorsell24-redis` | BuyOrSell24 | `buyorsell24_masaar_shared` | internal |
| `buyorsell24-ollama` | BuyOrSell24 | `buyorsell24_masaar_shared` | internal |
| `buyorsell24-api` | BuyOrSell24 | BuyOrSell24 network | internal (:8000) |
| `buyorsell24-app` | BuyOrSell24 | BuyOrSell24 network | internal (:9000) |
| `buyorsell24-frontend` | BuyOrSell24 | BuyOrSell24 network | internal (:3000) |
| `buyorsell24-worker` | BuyOrSell24 | BuyOrSell24 network | internal |
| `buyorsell24-meilisearch` | BuyOrSell24 | BuyOrSell24 network | internal (:7700) |
| `buyorsell24-mysql` | BuyOrSell24 | BuyOrSell24 network | internal |
| `masaar-api` | **Masaar CRM** | `buyorsell24_masaar_shared` | internal (:8080) |
| `masaar-web` | **Masaar CRM** | `buyorsell24_masaar_shared` | internal (:3000) |
| `masaar-postgres` | **Masaar CRM** | `buyorsell24_masaar_shared` | internal (:5432) |
| `dynamicweblab-web` | DynamicWebLab | `buyorsell24_masaar_shared` | internal (:3000) |
| `pause-frontend` | PausePOS | `pause_pause_internal` | internal (:3000) |
| `pause-backend` | PausePOS | `pause_pause_internal` | internal (:8080) |
| `pause-db` | PausePOS | `pause_pause_internal` | internal (:5432) |

### Masaar-specific shared services

| Service | Container | How Masaar uses it |
|---|---|---|
| **Reverse proxy** | `buyorsell24-nginx` | Routes `masaar.dynamicweblab.com` → `masaar-api:8080` + `masaar-web:3000` |
| **Redis** | `buyorsell24-redis` | Masaar uses **DB 1** (`redis://buyorsell24-redis:6379/1`) — BuyOrSell24 uses DB 0 |
| **Ollama LLM** | `buyorsell24-ollama` | Masaar calls it for PII AI tasks (same `buyorsell24_masaar_shared` network) |
| **SSL certs** | Host `/etc/letsencrypt` | Masaar cert: `/etc/letsencrypt/live/masaar.dynamicweblab.com/` |

### Nginx routing for Masaar (already configured in `buyorsell24-nginx`)

```
https://masaar.dynamicweblab.com/api/*    → masaar-api:8080
https://masaar.dynamicweblab.com/ws/*     → masaar-api:8080 (WebSocket)
https://masaar.dynamicweblab.com/webhooks/* → masaar-api:8080
https://masaar.dynamicweblab.com/health   → masaar-api:8080
https://masaar.dynamicweblab.com/*        → masaar-web:3000 (Next.js)
```

**Nginx config lives at:** `/root/projects/buyorsell24-engine/nginx.conf`  
(mounted read-only into `buyorsell24-nginx` — edit on host, then reload nginx)

---

## Method A — Deploy Code Changes (Git Pull + Rebuild)

Use after merging code to `main` on GitHub.

### 1. SSH into server

```bash
ssh -i /Users/maidul/info/contabo/backup-buyorsell24-key root@109.123.240.239
```

### 2. Pull latest code

```bash
cd /root/projects/masaar-crm
git pull origin main
```

### 3. Rebuild (run in background — survives SSH disconnects)

**API only (Go code changed):**
```bash
nohup docker compose -f docker-compose.server.yml build api > /tmp/masaar-api-build.log 2>&1 &
tail -f /tmp/masaar-api-build.log
```

**Web only (Next.js changed):**
```bash
nohup docker compose -f docker-compose.server.yml build web > /tmp/masaar-web-build.log 2>&1 &
tail -f /tmp/masaar-web-build.log
```

**Both:**
```bash
nohup docker compose -f docker-compose.server.yml build api web > /tmp/masaar-build.log 2>&1 &
tail -f /tmp/masaar-build.log
```

### 4. Restart Masaar containers only

```bash
cd /root/projects/masaar-crm
docker compose -f docker-compose.server.yml up -d api web
```

### 5. Verify

```bash
docker ps --format 'table {{.Names}}\t{{.Status}}' | grep masaar
docker logs masaar-api --tail 20
```

Expected output:
```
masaar-api    Up X minutes (healthy)
masaar-web    Up X minutes
masaar-postgres  Up X minutes (healthy)

AI cloud (non-PII): Gemini gemini-2.0-flash
SMSCountry integration enabled
Masaar CRM starting on :8080
```

---

## Method B — Update `.env` via SFTP (Config/Secret Changes)

Use when only environment variables changed — **no rebuild needed**, just restart.

### Option 1: SCP one-liner (fastest)

```bash
# Upload from Mac to server
scp -i /Users/maidul/info/contabo/backup-buyorsell24-key \
  /Users/maidul/projects/GitHub/Masaar-CRM/.env \
  root@109.123.240.239:/root/projects/masaar-crm/.env

# Then restart (no rebuild)
ssh -i /Users/maidul/info/contabo/backup-buyorsell24-key root@109.123.240.239 \
  "cd /root/projects/masaar-crm && docker compose -f docker-compose.server.yml up -d api web"
```

### Option 2: SFTP interactive

```bash
sftp -i /Users/maidul/info/contabo/backup-buyorsell24-key root@109.123.240.239
sftp> put /Users/maidul/projects/GitHub/Masaar-CRM/.env /root/projects/masaar-crm/.env
sftp> exit
```

Then restart via SSH (see Step 4 in Method A).

### Option 3: SFTP GUI (Cyberduck / Transmit / FileZilla)

| Field | Value |
|---|---|
| Protocol | SFTP |
| Host | `109.123.240.239` |
| Port | 22 |
| Username | `root` |
| Auth | Key file: `/Users/maidul/info/contabo/backup-buyorsell24-key` |
| Remote path | `/root/projects/masaar-crm` |

Upload `.env`, then restart containers via SSH.

---

## Method C — Update Nginx Config (Add/Change Routes)

The nginx config is **outside** the Masaar project — it belongs to BuyOrSell24 engine.

```bash
# Edit on server
ssh -i /Users/maidul/info/contabo/backup-buyorsell24-key root@109.123.240.239
nano /root/projects/buyorsell24-engine/nginx.conf

# Test config
docker exec buyorsell24-nginx nginx -t

# Reload nginx (zero downtime — does NOT restart the container)
docker exec buyorsell24-nginx nginx -s reload
```

> ⚠️ **Do not restart `buyorsell24-nginx`** unless absolutely necessary — it serves all projects.
=======
## Method B — Update `.env` via SFTP (Config/Secret Changes)

Use this when you only changed environment variables (API keys, tokens, secrets).

### Option 1: SFTP command line

**Both:**
```bash
sftp -i /Users/maidul/info/contabo/backup-buyorsell24-key root@109.123.240.239
```

Inside the SFTP session:
```
sftp> put /Users/maidul/projects/GitHub/Masaar-CRM/.env /root/projects/masaar-crm/.env
sftp> exit
```

Then SSH in and restart:
```bash
ssh -i /Users/maidul/info/contabo/backup-buyorsell24-key root@109.123.240.239
cd /root/projects/masaar-crm
docker compose -f docker-compose.server.yml up -d api web
```

### Option 2: SCP (one-liner, no interactive session)

```bash
scp -i /Users/maidul/info/contabo/backup-buyorsell24-key \
  /Users/maidul/projects/GitHub/Masaar-CRM/.env \
  root@109.123.240.239:/root/projects/masaar-crm/.env
```

Then restart:
```bash
ssh -i /Users/maidul/info/contabo/backup-buyorsell24-key root@109.123.240.239 \
  "cd /root/projects/masaar-crm && docker compose -f docker-compose.server.yml up -d api web"
```

### Option 3: SFTP GUI (Cyberduck / Transmit / FileZilla)

| Field | Value |
|---|---|
| Protocol | SFTP |
| Host | 109.123.240.239 |
| Port | 22 |
| Username | root |
| Auth | SSH key file: `/Users/maidul/info/contabo/backup-buyorsell24-key` |
| Remote path | `/root/projects/masaar-crm` |

Upload `.env` to `/root/projects/masaar-crm/.env`, then restart containers via SSH (Step 4 above).

---

## Method C — Upload a Single File via SFTP

Use this to push one file without a full git pull (e.g. hotfix a config).

```bash
sftp -i /Users/maidul/info/contabo/backup-buyorsell24-key root@109.123.240.239
```

```
# Upload a specific file (example: docker-compose.server.yml)
sftp> put /Users/maidul/projects/GitHub/Masaar-CRM/docker-compose.server.yml \
          /root/projects/masaar-crm/docker-compose.server.yml
sftp> exit
```

For code files — after uploading, always rebuild the affected container (Method A steps 3–5).

---

## What Changed — Current Production Stack

| Feature | Detail |
|---|---|
| **Redis** | Shared `buyorsell24-redis` on DB 1 (not a separate Masaar Redis) |
| **AI (PII)** | Ollama (local) — contacts, leads, WhatsApp messages never leave server |
| **AI (public)** | Gemini `gemini-2.0-flash` — property listing copy only |
| **Email** | Azure Communication Services (`buyorsell24-communication.uae.communication.azure.com`) |
| **SMS OTP** | SMSCountry — phone login/registration, auto-creates Viewer accounts |
| **Magic Link** | Token emailed via Azure, lands on `/login/magic-link/verify?token=...` |
| **Forgot Password** | `/forgot-password` and `/reset-password` pages wired to backend |
| **BOS24** | UAE real estate data from `https://data.buyorsell24.com` |
| **Network** | `buyorsell24_masaar_shared` external Docker network |

---

## Current `.env` Key Values (Production)

> Full `.env` is at `/Users/maidul/projects/GitHub/Masaar-CRM/.env`  
> Never commit `.env` to git.

| Variable | Where to find it |
|---|---|
| `DATABASE_URL` | Postgres on `masaar-postgres:5432` |
| `REDIS_URL` | `redis://buyorsell24-redis:6379/1` (DB 1) |
| `JWT_SECRET` | 64-char hex in `.env` |
| `GEMINI_API_KEY` | Google AI Studio |
| `AZURE_COMM_KEY` | Azure portal → Communication Services |
| `SMSCOUNTRY_AUTH_KEY/TOKEN` | SMSCountry dashboard |
| `BOS24_API_TOKEN` | BuyOrSell24 admin |
>>>>>>> b78c68cc07c568270c9d85ed9c6c39668b24debd

---

## Deployment Decision Tree

```
<<<<<<< HEAD
Changed .env only?
  └─ YES → Method B (SCP .env → restart api web — no rebuild)

Changed Go backend code?
  └─ YES → Method A (git pull → build api → up -d api web)

Changed Next.js frontend?
  └─ YES → Method A (git pull → build web → up -d api web)

Changed nginx routing?
  └─ YES → Method C (edit nginx.conf → nginx -t → nginx -s reload)
=======
Did you change .env only?
  └─ YES → Method B (SFTP/SCP .env, restart containers — no rebuild needed)
  └─ NO  → Did you change Go backend code?
              └─ YES → Method A (git pull → build api → up -d)
              └─ NO  → Did you change Next.js frontend?
                          └─ YES → Method A (git pull → build web → up -d)
>>>>>>> b78c68cc07c568270c9d85ed9c6c39668b24debd
```

---

## Health Checks
<<<<<<< HEAD

```bash
# Container status
docker ps --format 'table {{.Names}}\t{{.Status}}' | grep -E 'masaar|NAME'

# API live check
curl -s https://masaar.dynamicweblab.com/health

# API startup logs
docker logs masaar-api --tail 30

# Database alive
docker exec masaar-postgres psql -U masaar -d masaar -c "SELECT 1;"

# Redis alive — Masaar's DB 1
docker exec buyorsell24-redis redis-cli -n 1 PING

# Ollama alive (shared container)
curl -s http://localhost:11434/api/tags | python3 -m json.tool 2>/dev/null | grep name
=======

```bash
# All containers
docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'

# API health endpoint
curl -s https://masaar.dynamicweblab.com/health

# API startup logs
docker logs masaar-api --tail 30

# Database alive
docker exec masaar-postgres psql -U masaar -d masaar -c "SELECT 1;"

# Redis alive (DB 1)
docker exec buyorsell24-redis redis-cli -n 1 PING

# Ollama alive
curl -s http://localhost:11434/api/tags | head -c 200
```

---

## Rollback

If the new build breaks something:

```bash
ssh -i /Users/maidul/info/contabo/backup-buyorsell24-key root@109.123.240.239
cd /root/projects/masaar-crm

# Roll back to previous image (Docker keeps last image)
docker compose -f docker-compose.server.yml up -d --no-build api web

# OR roll back git and rebuild
git log --oneline -5          # find the good commit
git checkout <commit-hash>
nohup docker compose -f docker-compose.server.yml build api web > /tmp/rollback.log 2>&1 &
tail -f /tmp/rollback.log
# then: docker compose -f docker-compose.server.yml up -d api web
>>>>>>> b78c68cc07c568270c9d85ed9c6c39668b24debd
```

---

<<<<<<< HEAD
## Rollback

```bash
ssh -i /Users/maidul/info/contabo/backup-buyorsell24-key root@109.123.240.239
cd /root/projects/masaar-crm

# Option 1: Roll back git commit and rebuild
git log --oneline -5
git checkout <good-commit-hash>
nohup docker compose -f docker-compose.server.yml build api web > /tmp/rollback.log 2>&1 &
tail -f /tmp/rollback.log
docker compose -f docker-compose.server.yml up -d api web

# Option 2: If Docker kept the previous image layer (quick)
docker compose -f docker-compose.server.yml up -d --no-build api web
=======
## Database Backup

```bash
# On the server — dump to local file
docker exec masaar-postgres pg_dump -U masaar masaar \
  | gzip > /root/backups/masaar-$(date +%Y%m%d-%H%M).sql.gz

# Pull backup to your Mac via SFTP
sftp -i /Users/maidul/info/contabo/backup-buyorsell24-key root@109.123.240.239
sftp> get /root/backups/masaar-*.sql.gz /Users/maidul/Downloads/
>>>>>>> b78c68cc07c568270c9d85ed9c6c39668b24debd
```

---

<<<<<<< HEAD
## Database Backup

```bash
# On server
docker exec masaar-postgres pg_dump -U masaar masaar \
  | gzip > /root/backups/masaar-$(date +%Y%m%d-%H%M).sql.gz

# Pull backup to your Mac
scp -i /Users/maidul/info/contabo/backup-buyorsell24-key \
  "root@109.123.240.239:/root/backups/masaar-*.sql.gz" \
  ~/Downloads/
```

---

## Logs

```bash
# Live API logs
docker logs masaar-api -f

=======
## Logs

```bash
# Live API logs
docker logs masaar-api -f

>>>>>>> b78c68cc07c568270c9d85ed9c6c39668b24debd
# Last 50 lines
docker logs masaar-api --tail 50

# Web logs
<<<<<<< HEAD
docker logs masaar-web --tail 30

# Nginx access logs (all sites)
docker logs buyorsell24-nginx --tail 50

# Build output
tail -f /tmp/masaar-build.log
=======
docker logs masaar-web --tail 50

# Build output (if build was run with nohup)
tail -f /tmp/masaar-build.log
tail -f /tmp/masaar-web-build.log
>>>>>>> b78c68cc07c568270c9d85ed9c6c39668b24debd
```

---

<<<<<<< HEAD
## Current Stack Summary

| Component | Detail |
|---|---|
| **Domain** | `masaar.dynamicweblab.com` |
| **SSL** | Let's Encrypt cert, terminated at `buyorsell24-nginx` |
| **Redis** | `buyorsell24-redis:6379` DB 1 (shared, DB 0 = BuyOrSell24) |
| **AI — PII** | `buyorsell24-ollama:11434` — contacts, leads, WhatsApp (stays local, UAE PDPL) |
| **AI — Public** | Gemini `gemini-2.0-flash` — property listing copy only |
| **Email** | Azure Communication Services (UAE region) |
| **SMS OTP** | SMSCountry — phone login, auto-creates Viewer accounts |
| **Real estate data** | BuyOrSell24 API `https://data.buyorsell24.com` |
| **Docker network** | `buyorsell24_masaar_shared` (external, shared with BuyOrSell24 engine) |

---

**Last updated:** 2026-05-23  
**Server:** `109.123.240.239` · `/root/projects/masaar-crm` · `docker-compose.server.yml`
=======
**Last updated:** 2026-05-23  
**Server:** 109.123.240.239 · `/root/projects/masaar-crm` · `docker-compose.server.yml`
>>>>>>> b78c68cc07c568270c9d85ed9c6c39668b24debd
