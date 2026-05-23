# Masaar CRM — Production Deployment Guide

**Live URL:** https://masaar.dynamicweblab.com  
**Server:** 109.123.240.239 (Contabo VPS, Ubuntu 24.04)  
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
```

---

## Method B — Update `.env` via SFTP (Config/Secret Changes)

Use this when you only changed environment variables (API keys, tokens, secrets).

### Option 1: SFTP command line

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

---

## Deployment Decision Tree

```
Did you change .env only?
  └─ YES → Method B (SFTP/SCP .env, restart containers — no rebuild needed)
  └─ NO  → Did you change Go backend code?
              └─ YES → Method A (git pull → build api → up -d)
              └─ NO  → Did you change Next.js frontend?
                          └─ YES → Method A (git pull → build web → up -d)
```

---

## Health Checks

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
```

---

## Database Backup

```bash
# On the server — dump to local file
docker exec masaar-postgres pg_dump -U masaar masaar \
  | gzip > /root/backups/masaar-$(date +%Y%m%d-%H%M).sql.gz

# Pull backup to your Mac via SFTP
sftp -i /Users/maidul/info/contabo/backup-buyorsell24-key root@109.123.240.239
sftp> get /root/backups/masaar-*.sql.gz /Users/maidul/Downloads/
```

---

## Logs

```bash
# Live API logs
docker logs masaar-api -f

# Last 50 lines
docker logs masaar-api --tail 50

# Web logs
docker logs masaar-web --tail 50

# Build output (if build was run with nohup)
tail -f /tmp/masaar-build.log
tail -f /tmp/masaar-web-build.log
```

---

**Last updated:** 2026-05-23  
**Server:** 109.123.240.239 · `/root/projects/masaar-crm` · `docker-compose.server.yml`
