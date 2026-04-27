# Masaar CRM — Production Deployment Checklist

## Pre-Deployment (NOW)

- [ ] Review `DEPLOYMENT.md` completely
- [ ] Decide on domains (e.g., api.example.com, crm.example.com)
- [ ] Generate secure JWT_SECRET: `openssl rand -base64 32`
- [ ] Generate secure DB password
- [ ] Generate secure Redis password
- [ ] Get WhatsApp Business API credentials (optional for MVP)
- [ ] Prepare SMTP credentials (optional)

## Deployment Day

### 1. Server Preparation (5 min)
```bash
ssh root@109.123.240.239

# Create deployment directory
mkdir -p /opt/masaar-crm
cd /opt/masaar-crm

# Optional: Clean up old Docker images (frees ~20GB)
docker system prune -a --force
```

- [ ] Connected to server
- [ ] Created /opt/masaar-crm directory
- [ ] Disk space freed (if needed)

### 2. Clone & Configure (10 min)
```bash
# Clone repository
git clone https://github.com/maidulcu/masaar-crm.git .

# Copy and configure
cp .env.example .env
nano .env  # Edit with YOUR values
```

Update these in `.env`:
- [ ] `DATABASE_URL` — Change password from "masaar" to secure value
- [ ] `REDIS_URL` — Change to `redis://redis:6380` and set password
- [ ] `JWT_SECRET` — Set to `openssl rand -base64 32` output
- [ ] `WA_VERIFY_TOKEN` — Set a random webhook token (you'll use this in Meta)
- [ ] `NEXT_PUBLIC_API_URL` — Will set to `https://api.yourdomain.com`

### 3. Docker Compose Setup (5 min)
```bash
# Update docker-compose.yml with custom ports and passwords
# Copy template from DEPLOYMENT.md Step 4
nano docker-compose.yml
```

- [ ] docker-compose.yml updated with production settings
- [ ] Volume directories created (`mkdir -p volumes/{postgres,redis,ollama}`)

### 4. Start Services (15-20 min first run)
```bash
docker compose up -d

# Watch Ollama downloading model (takes 5-10 min)
docker compose logs ollama -f

# When model is done, verify all services healthy
docker compose ps
```

Expected status:
```
masaar-postgres  Up (healthy)
masaar-redis     Up (healthy)
masaar-ollama    Up (when model done)
masaar-api       Up (healthy)
masaar-web       Up
```

- [ ] All containers started
- [ ] Ollama model downloaded successfully
- [ ] API is responding (curl http://localhost:8080/api/v1/stats)

### 5. Nginx Configuration (5 min)
```bash
# Update nginx to route traffic
sudo nano /etc/nginx/sites-available/default
# Add upstream blocks and server blocks from DEPLOYMENT.md Step 6

# Test and reload
sudo nginx -t
sudo systemctl reload nginx
```

- [ ] Nginx config updated
- [ ] Nginx test passed
- [ ] Nginx reloaded

### 6. DNS Setup (You Handle)

Point your DNS to: **109.123.240.239**
```
A Record: api.yourdomain.com      → 109.123.240.239
A Record: crm.yourdomain.com      → 109.123.240.239
```

- [ ] DNS records created
- [ ] DNS propagated (check: `nslookup api.yourdomain.com`)

### 7. SSL/TLS Setup (5-10 min)
```bash
sudo certbot certonly --nginx -d api.yourdomain.com -d crm.yourdomain.com

# Update nginx with SSL certs from DEPLOYMENT.md Step 7
sudo nano /etc/nginx/sites-available/default
sudo nginx -t
sudo systemctl reload nginx
```

- [ ] SSL certificates obtained
- [ ] Nginx updated with SSL config
- [ ] HTTPS working (`curl https://api.yourdomain.com`)

### 8. Post-Deployment Tests (5 min)

```bash
# Test API
curl https://api.yourdomain.com/api/v1/stats

# Test Frontend
curl https://crm.yourdomain.com

# Test Database
docker exec masaar-postgres psql -U masaar -d masaar -c "SELECT version();"

# Test Redis
docker exec masaar-redis redis-cli -p 6380 PING

# Test Ollama
curl http://localhost:11434/api/tags
```

- [ ] API responding (port 8080)
- [ ] Frontend loading (port 3000)
- [ ] Database healthy
- [ ] Redis healthy
- [ ] Ollama ready

### 9. Security Lockdown (10 min)

```bash
# Change default admin password
# Login via UI at https://crm.yourdomain.com
# Email: admin@masaar.local
# Password: changeme
# → Go to Settings and change password immediately
```

- [ ] Default password changed
- [ ] JWT_SECRET is random (not default)
- [ ] Database password secured
- [ ] Redis password secured
- [ ] HTTPS enforced (redirect HTTP → HTTPS)

### 10. Setup Monitoring & Backups (10 min)

```bash
# Create backup script
mkdir -p /opt/masaar-crm/backups
chmod 755 /opt/masaar-crm/backups

# Add backup cron (see DEPLOYMENT.md)
crontab -e
# 0 2 * * * /opt/masaar-crm/backup.sh

# Setup health monitoring
docker compose logs -f &  # Optional: keep monitoring
```

- [ ] Backup directory created
- [ ] Backup cron job scheduled
- [ ] Monitoring set up

---

## Post-Deployment (Next 24 Hours)

- [ ] Monitor logs: `docker compose logs --tail 100`
- [ ] Test WhatsApp webhook (if you have credentials)
- [ ] Create test lead/contact via UI
- [ ] Verify email works (if SMTP configured)
- [ ] Check disk space: `df -h /`
- [ ] Verify automated backups run

---

## Rollback Plan (If Needed)

```bash
# Stop everything
docker compose down

# Restore from backup
gunzip < backups/db-YYYYMMDD-HHMMSS.sql.gz | \
  docker exec -i masaar-postgres psql -U masaar masaar

# Restart
docker compose up -d
```

---

## Estimated Timeline

| Step | Duration | Cumulative |
|------|----------|-----------|
| 1. Prep | 5 min | 5 min |
| 2. Clone & Config | 10 min | 15 min |
| 3. Docker Setup | 5 min | 20 min |
| 4. Start Services | 20 min | **40 min** |
| 5. Nginx Config | 5 min | 45 min |
| 6. DNS (async) | 24-48h | — |
| 7. SSL Setup | 10 min | 55 min |
| 8. Post-Tests | 5 min | **60 min** |
| 9. Security | 10 min | 70 min |
| 10. Monitoring | 10 min | **80 min** |

**Total:** ~80 minutes (excluding DNS propagation time)

---

## Support & Debugging

### If something fails:

1. **Check logs:** `docker compose logs [service] --tail 50`
2. **Check ports:** `netstat -tulpn | grep -E "8080|3000|5432|6380"`
3. **Check disk:** `df -h /` (ensure > 50GB free)
4. **Check memory:** `free -h` (should have > 10GB free)
5. **Restart service:** `docker compose restart [service]`
6. **Full reset:** `docker compose down && docker compose up -d`

### Common Issues:

- **Ollama stuck:** Check logs with `docker compose logs ollama -f`, may take 20+ min for first run
- **API won't start:** Check database health first (`docker compose logs postgres`)
- **Frontend blank:** Check API URL is correct in environment
- **Port conflicts:** Use `lsof -i :PORT` to find what's using it

---

**You're ready! Start with Step 1 when DNS is set up.** 🚀
