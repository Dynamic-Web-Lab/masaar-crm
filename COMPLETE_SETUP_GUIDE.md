# 🎯 Complete Masaar CRM Production Setup Guide

Everything you need for **instant-deployment production setup** with git automation.

---

## 📦 Complete File Inventory

### Deployment Files (7)
```
DEPLOYMENT_README.md          Quick reference & overview
DEPLOYMENT.md                 Complete 10-step deployment guide
DEPLOYMENT_CHECKLIST.md       Step-by-step checklist (80 min)
deploy.sh (executable)        Auto-setup script
docker-compose.prod.override.yml  Production Docker config
nginx.conf.template           Reverse proxy config
OPERATIONS.md                 Daily operations & troubleshooting
```

### Git Automation Files (3) ✨ NEW
```
GIT_AUTOMATION.md             Complete git automation guide
setup-git-automation.sh       One-command hook setup
git-hooks/post-receive        Auto-deploy hook
```

**Total:** 10 production-ready files, all committed ✅

---

## 🚀 Complete Deployment Workflow

### Phase 1: Initial Setup (45 min active)

```bash
# 1. SSH to server
ssh root@109.123.240.239

# 2. Run automated deployment
cd /opt/masaar-crm
sudo bash deploy.sh

# Follows DEPLOYMENT_CHECKLIST.md for remaining steps:
# - Edit .env with secrets
# - Setup Nginx
# - Get SSL certificate
# - Update DNS
```

### Phase 2: Setup Git Automation (2 min)

```bash
# After initial deployment is successful:
sudo bash setup-git-automation.sh

# Done! Now every push auto-deploys
```

### Phase 3: Production Ready 🎉

```bash
# Deploy code with a simple push
git push origin main

# Automatic deployment happens on server:
# - Code pulled
# - Docker images rebuilt
# - Containers restarted
# - Health checks run
# All within ~50 seconds!
```

---

## 📋 Step-by-Step Guide

### 1️⃣ Initial Deployment (Do Once)

**Read:** `DEPLOYMENT_README.md` (overview)
**Follow:** `DEPLOYMENT_CHECKLIST.md` (step-by-step)
**Reference:** `DEPLOYMENT.md` (detailed)

**Commands on Server:**
```bash
ssh root@109.123.240.239
cd /opt/masaar-crm
sudo bash deploy.sh
# Edit .env
# Setup Nginx
# Get SSL cert
```

**Time:** ~45 minutes active work

### 2️⃣ Setup Git Automation (Do Once)

**Read:** `GIT_AUTOMATION.md` (how it works)

**Commands on Server:**
```bash
sudo bash setup-git-automation.sh
# That's it! Auto-deploy is now active
```

**Time:** ~2 minutes

### 3️⃣ Daily Development (Ongoing)

**Deploy new code:**
```bash
git push origin main
# Server auto-deploys within 50 seconds
```

**Monitor deployment:**
```bash
ssh root@109.123.240.239
tail -f /var/log/masaar-deploy.log
```

**Daily operations:**
```bash
# See OPERATIONS.md for:
# - Container management
# - Database backups
# - Troubleshooting
# - Health monitoring
```

---

## 🎯 Key Features

### Deployment
✅ One-command automated setup (`deploy.sh`)
✅ Production Docker configuration ready
✅ Nginx reverse proxy template included
✅ Let's Encrypt SSL setup instructions
✅ Health checks on every deploy
✅ Rollback procedures documented

### Git Automation
✅ Auto-deploy on `git push origin main`
✅ Code pulled automatically
✅ Docker images rebuilt
✅ Containers restarted
✅ Health checks run
✅ Logs saved automatically
✅ ~50 second deployment time

### Operations
✅ Daily container management commands
✅ Database backup procedures
✅ Troubleshooting guides
✅ Performance monitoring
✅ Security checklists
✅ Emergency procedures

---

## 🖥️ Your Server Status

| Item | Status | Details |
|------|--------|---------|
| OS | ✅ Ubuntu 24.04 LTS | Perfect |
| CPU | ✅ 12 cores | Excellent |
| RAM | ✅ 47 GB | Excellent |
| Disk | ✅ 75 GB free | Sufficient |
| Docker | ✅ v29.4.0 | Ready |
| Docker Compose | ✅ v5.1.3 | Ready |
| Git | ✅ Available | Ready for automation |
| Existing Services | ⚠️ BuyOrSell24 | Won't conflict (different ports) |

---

## 📚 Documentation Map

```
For First-Time Setup:
  1. DEPLOYMENT_README.md          (orientation)
  2. DEPLOYMENT_CHECKLIST.md       (follow step-by-step)
  3. DEPLOYMENT.md                 (detailed reference)

For Git Automation Setup:
  1. GIT_AUTOMATION.md             (read how it works)
  2. setup-git-automation.sh       (run once)

For Daily Operations:
  1. OPERATIONS.md                 (all daily tasks)
  
For Troubleshooting:
  1. OPERATIONS.md → Troubleshooting section
  2. DEPLOYMENT.md → Troubleshooting section
```

---

## ⚡ Quick Commands Reference

### Deploy & Monitor
```bash
git push origin main                    # Deploy
tail -f /var/log/masaar-deploy.log     # Watch deployment
docker compose ps                       # Check status
```

### Development
```bash
docker compose logs -f api              # API logs
docker compose logs -f web              # Frontend logs
curl http://localhost:8080/api/v1/stats # Health check
```

### Database
```bash
docker exec -it masaar-postgres psql -U masaar -d masaar
pg_dump -U masaar masaar | gzip > backup.sql.gz
```

### Troubleshooting
```bash
docker compose down && docker compose up -d   # Full restart
docker system prune -a                        # Free disk space
tail -100 /var/log/masaar-deploy.log         # Check deployment log
```

---

## 🔐 Security Setup Checklist

Before going live:
- [ ] Changed JWT_SECRET (run: `openssl rand -base64 32`)
- [ ] Changed POSTGRES_PASSWORD (not "masaar")
- [ ] Changed REDIS_PASSWORD
- [ ] Changed default admin password
- [ ] Enabled HTTPS/SSL
- [ ] Set WhatsApp webhook token
- [ ] Configured firewall (only 80, 443)
- [ ] Setup automated backups
- [ ] Reviewed OPERATIONS.md security section

---

## 🎓 Learning Path

### Day 1: Deploy
1. Read `DEPLOYMENT_README.md` (5 min)
2. Prepare domain & secrets (10 min)
3. Follow `DEPLOYMENT_CHECKLIST.md` (45 min)
4. Verify everything works (10 min)

### Day 2: Git Automation
1. Read `GIT_AUTOMATION.md` (10 min)
2. Run `setup-git-automation.sh` (2 min)
3. Test git deployment (5 min)

### Day 3+: Operations
1. Read `OPERATIONS.md` sections as needed
2. Schedule backups
3. Monitor logs
4. Keep deploying! 🚀

---

## 📊 Timeline Summary

| Phase | Duration | What Happens |
|-------|----------|--------------|
| **Initial Deploy** | 45 min | Server setup, Docker start, Nginx config |
| **DNS Propagation** | 24-48h | (automatic, user doesn't do anything) |
| **Git Automation** | 2 min | Setup auto-deploy hooks |
| **Production Ready** | 2-3 days | Complete setup + DNS propagation |

---

## ❓ Common Questions

### Q: What if I push code before git automation is set up?
**A:** It won't auto-deploy. You need to run `setup-git-automation.sh` first.

### Q: Can I rollback if a deployment fails?
**A:** Yes! Push the previous commit: `git push -f origin main`
Or manually restart: `docker compose restart`

### Q: How long does deployment take?
**A:** ~50 seconds for typical changes. Longer builds take more time.

### Q: What gets deployed?
**A:** Backend code, frontend code, migrations. NOT .env, docker-compose.yml, or data.

### Q: Can I deploy other branches?
**A:** Hook only deploys 'main'. Push main to deploy: `git push origin main`

### Q: Where are deployment logs?
**A:** Server: `/var/log/masaar-deploy.log`
Command: `tail -f /var/log/masaar-deploy.log`

---

## 🚨 Emergency Procedures

### If deployment fails:

```bash
# Check logs
tail -100 /var/log/masaar-deploy.log

# Manual recovery
cd /opt/masaar-crm
git status
docker compose logs --tail 50
docker compose restart

# Full rollback (push previous commit)
git reset --soft HEAD~1
git push -f origin main
```

### If server is down:

```bash
# SSH to server
ssh root@109.123.240.239

# Restart everything
cd /opt/masaar-crm
docker compose down
docker compose up -d

# Verify health
curl http://localhost:8080/api/v1/stats
```

### If disk is full:

```bash
# Free up space
docker system prune -a --force
rm -rf /opt/masaar-crm/backups/*.sql.gz

# Check space
df -h /
```

---

## ✅ Verification Checklist

After complete setup, verify:

```bash
# API health
curl https://api.yourdomain.com/api/v1/stats → 200

# Frontend loads
curl https://crm.yourdomain.com → HTML response

# Database works
docker exec masaar-postgres psql -U masaar -c "SELECT 1;" → 1

# Redis works
docker exec masaar-redis redis-cli -p 6380 PING → PONG

# Git automation ready
ls -x /opt/masaar-crm/.git/hooks/post-receive → exists & executable

# Logs available
cat /var/log/masaar-deploy.log → shows deployment history
```

---

## 📞 Support Quick Links

| Issue | Solution |
|-------|----------|
| Deployment failed | Check `/var/log/masaar-deploy.log` |
| API not responding | See OPERATIONS.md → API Health Checks |
| Database issues | See OPERATIONS.md → Database Operations |
| Disk full | See OPERATIONS.md → Performance Monitoring |
| Rollback needed | Push previous commit: `git push -f origin main` |
| Setup help | See DEPLOYMENT.md → Step X |
| Git automation help | See GIT_AUTOMATION.md |

---

## 🎉 You're All Set!

Everything is ready for production deployment with git automation.

**Next steps:**
1. Read `DEPLOYMENT_README.md`
2. Follow `DEPLOYMENT_CHECKLIST.md`
3. Run `setup-git-automation.sh`
4. Deploy with `git push origin main`

**Happy deploying!** 🚀

---

**Branch:** `claude/mvp-server-readiness-eNKuU`
**Files:** 10 production-ready configs & guides
**Status:** ✅ Ready for deployment
**Last updated:** 2026-04-27
