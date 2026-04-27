# 🚀 Masaar CRM — Production Deployment Guide

**Everything you need to deploy Masaar CRM to your production server.**

---

## 📋 Files Created for Deployment

| File | Size | Purpose |
|------|------|---------|
| **DEPLOYMENT.md** | 13K | Complete deployment guide with all steps |
| **DEPLOYMENT_CHECKLIST.md** | 6.3K | Step-by-step checklist to follow |
| **deploy.sh** | 4.7K | Automated deployment script (run on server) |
| **docker-compose.prod.override.yml** | - | Production Docker configuration |
| **nginx.conf.template** | - | Reverse proxy configuration template |
| **OPERATIONS.md** | 8.6K | Day-to-day operations and troubleshooting |

---

## ⚡ Quick Start (5 minutes)

### 1. Copy Files to Production Server

```bash
# On your local machine:
scp -r /path/to/masaar-crm/* root@109.123.240.239:/opt/masaar-crm/
```

Or clone directly on server:
```bash
# On the server:
cd /opt
git clone https://github.com/maidulcu/masaar-crm.git masaar-crm
cd masaar-crm
```

### 2. Run Deployment Script

```bash
# SSH into server
ssh root@109.123.240.239

# Run automated deployment
cd /opt/masaar-crm
sudo bash deploy.sh
```

The script will:
- ✅ Clone/update repository
- ✅ Create `.env` file (you'll edit it)
- ✅ Set up Docker volumes
- ✅ Start all containers
- ✅ Run health checks

### 3. Configure Environment

Edit `.env` with your production values:
```bash
nano /opt/masaar-crm/.env
```

**Must change:**
- `JWT_SECRET` → `openssl rand -base64 32`
- `POSTGRES_PASSWORD` → Secure password
- `REDIS_PASSWORD` → Secure password
- `WA_VERIFY_TOKEN` → Your webhook token (from Meta)
- `NEXT_PUBLIC_API_URL` → `https://api.yourdomain.com`

### 4. Setup Nginx Reverse Proxy

```bash
# Copy nginx config template
sudo cp /opt/masaar-crm/nginx.conf.template /etc/nginx/sites-available/default

# Edit with your domain
sudo nano /etc/nginx/sites-available/default

# Test and reload
sudo nginx -t
sudo systemctl reload nginx
```

### 5. Get SSL Certificate

```bash
# Install Let's Encrypt
sudo apt-get install certbot python3-certbot-nginx

# Get certificate
sudo certbot certonly --nginx -d api.yourdomain.com -d crm.yourdomain.com

# Update nginx config with SSL paths
sudo nano /etc/nginx/sites-available/default
sudo nginx -t && sudo systemctl reload nginx
```

### 6. Update DNS

Point your DNS records to **109.123.240.239**:
```
A Record: api.yourdomain.com      → 109.123.240.239
A Record: crm.yourdomain.com      → 109.123.240.239
```

### 7. Verify Deployment

```bash
# Check container status
docker compose ps

# Test API
curl https://api.yourdomain.com/api/v1/stats

# Test Frontend
curl https://crm.yourdomain.com
```

---

## 📖 Documentation Structure

### For Deployment
1. **Start here:** `DEPLOYMENT_CHECKLIST.md` — Step-by-step checklist
2. **Detailed steps:** `DEPLOYMENT.md` — Full deployment guide with explanations
3. **Automation:** `deploy.sh` — Run this first to automate setup

### For Daily Operations
- **Operations:** `OPERATIONS.md` — Container management, backups, troubleshooting

### Configuration Templates
- **Nginx config:** `nginx.conf.template` — Reverse proxy setup
- **Docker config:** `docker-compose.prod.override.yml` — Production Docker settings

---

## 🎯 Recommended Deployment Path

### Phase 1: Initial Deployment (Day 1)
```
1. SSH to server
2. Run: sudo bash deploy.sh
3. Edit .env with your secrets
4. Wait for containers to start (10-15 min)
5. Setup Nginx
6. Get SSL certificate
```
**Time: ~1-2 hours**

### Phase 2: Configuration (Day 2)
```
7. Update DNS records
8. Wait for DNS propagation (24-48 hours)
9. Change default admin password
10. Test WhatsApp webhook (if credentials ready)
11. Configure SMTP (optional)
```
**Time: ~30 minutes active, 24-48 hours waiting**

### Phase 3: Monitoring (Ongoing)
```
- Daily: Check logs and health
- Weekly: Backup database
- Monthly: Review security and disk space
```

---

## 🔐 Security Checklist

Before going live, ensure:
- [ ] `JWT_SECRET` is randomly generated (not from .env.example)
- [ ] `POSTGRES_PASSWORD` is changed from "masaar"
- [ ] `REDIS_PASSWORD` is set to a secure value
- [ ] Default admin password changed (admin@masaar.local / changeme)
- [ ] SSL/TLS enabled (HTTPS only)
- [ ] WhatsApp webhook token set (if using WhatsApp)
- [ ] Firewall configured (only ports 80, 443 exposed)
- [ ] Regular backups scheduled (cron job)

---

## 📊 Server Requirements

Your server (vmi3243176) meets all requirements:

| Requirement | Your Server | Status |
|-------------|------------|--------|
| **CPU** | 12 cores | ✅ Excellent |
| **RAM** | 47 GB | ✅ Excellent |
| **Disk** | 75 GB free | ✅ Sufficient |
| **OS** | Ubuntu 24.04 LTS | ✅ Perfect |
| **Docker** | v29.4.0 | ✅ Ready |
| **Docker Compose** | v5.1.3 | ✅ Ready |

---

## 🚨 Troubleshooting Quick Links

| Issue | Solution |
|-------|----------|
| Ollama stuck downloading | See OPERATIONS.md → Troubleshooting |
| API won't start | See OPERATIONS.md → Database Operations |
| Frontend blank | See OPERATIONS.md → API Health Checks |
| Port conflicts | See OPERATIONS.md → Port Management |
| Disk space full | See OPERATIONS.md → Performance Monitoring |

---

## 📞 Getting Help

1. **Before asking for help:**
   - Check container logs: `docker compose logs --tail 100`
   - Check disk space: `df -h /`
   - Check memory: `free -h`

2. **If something fails:**
   - See `DEPLOYMENT.md` → Troubleshooting
   - See `OPERATIONS.md` → Troubleshooting
   - Check `.env` configuration (most common issue)

3. **For specific commands:**
   - See `OPERATIONS.md` for quick reference table

---

## 📋 Next Steps

1. **Prepare your production server details:**
   - Domain name (e.g., api.masaar.example.com)
   - SSL certificate provider (Let's Encrypt)
   - Database backup location
   - Admin email/password

2. **SSH to your server:**
   ```bash
   ssh root@109.123.240.239
   cd /opt/masaar-crm
   ```

3. **Run the deployment script:**
   ```bash
   sudo bash deploy.sh
   ```

4. **Follow DEPLOYMENT_CHECKLIST.md** for each step

---

## 📌 Important Notes

- **First run takes longer:** Ollama model download can take 10-20 minutes
- **DNS propagation:** May take 24-48 hours for DNS to propagate globally
- **Backup early:** Create database backup before making user accounts
- **Monitor resources:** Server is at 70% disk usage; keep an eye on it

---

**Questions? See DEPLOYMENT.md or OPERATIONS.md**

**Last updated:** 2026-04-27
