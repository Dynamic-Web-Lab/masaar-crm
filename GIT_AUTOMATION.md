# Masaar CRM — Git Automation & CI/CD

Automatic deployments when you push code to the repository.

---

## ⚡ Quick Setup (2 minutes)

After initial deployment, setup auto-deployments:

```bash
# SSH to server
ssh root@109.123.240.239

# Run setup script (installs git hooks)
cd /opt/masaar-crm
sudo bash setup-git-automation.sh

# Done! Every push to 'main' now auto-deploys
```

---

## 🔄 How It Works

### Workflow
```
Developer pushes code
       ↓
Server receives push
       ↓
Git post-receive hook triggers
       ↓
Script pulls latest code
       ↓
Docker images rebuilt
       ↓
Containers restarted
       ↓
Health checks run
       ↓
Logs written to /var/log/masaar-deploy.log
```

### Timeline
- **0-5 sec:** Code pulled
- **5-30 sec:** Docker images built (depends on changes)
- **30-40 sec:** Containers stopped and started
- **40-50 sec:** Health checks
- **Total:** ~50 seconds for typical deploy

---

## 📤 Deploying Code

### From Your Local Machine

```bash
# Make changes locally
git add .
git commit -m "Feature: Add new contact field"

# Push to server (triggers auto-deploy)
git push origin main

# Server auto-deploys! Monitor logs:
ssh root@109.123.240.239
tail -f /var/log/masaar-deploy.log
```

### What Gets Deployed
- ✅ Backend code changes
- ✅ Frontend code changes
- ✅ Database migrations (run automatically on startup)
- ✅ Environment variables (from .env on server)

### What Does NOT Get Deployed
- ❌ Changes to .env file (edit directly on server)
- ❌ Changes to docker-compose.yml (update manually)
- ❌ Database data (preserved between deployments)

---

## 📊 Monitoring Deployments

### Watch Live Deployment

```bash
# SSH to server
ssh root@109.123.240.239

# Watch deployment logs in real-time
tail -f /var/log/masaar-deploy.log

# In another terminal, check container status
docker compose ps

# Check API health
curl http://localhost:8080/api/v1/stats
```

### View Deployment History

```bash
# See all deployments
cat /var/log/masaar-deploy.log

# See last 50 deployments
tail -50 /var/log/masaar-deploy.log

# See specific deployment (search by timestamp)
grep "2026-04-27 15:30" /var/log/masaar-deploy.log
```

### View Container Logs

```bash
# API logs
docker compose logs api --tail 50

# Frontend logs
docker compose logs web --tail 50

# Database logs
docker compose logs postgres --tail 50

# Follow logs in real-time
docker compose logs -f
```

---

## 🛠️ Installation Details

### What the Setup Script Does

1. **Creates `.git/hooks/post-receive`**
   - Runs automatically when code is pushed
   - Pulls latest code
   - Rebuilds Docker images
   - Restarts containers
   - Logs to `/var/log/masaar-deploy.log`

2. **Sets up deployment logging**
   - All deployments logged with timestamps
   - Useful for debugging failed deployments
   - Can be rotated to save disk space

3. **Verifies git configuration**
   - Ensures 'origin' remote is configured
   - Confirms branch tracking
   - Shows last commit info

### Manual Installation (if needed)

```bash
# 1. SSH to server
ssh root@109.123.240.239

# 2. Create hooks directory
mkdir -p /opt/masaar-crm/.git/hooks

# 3. Create post-receive hook
cat > /opt/masaar-crm/.git/hooks/post-receive << 'EOF'
#!/bin/bash
# (copy entire hook script from git-hooks/post-receive)
EOF

# 4. Make it executable
chmod +x /opt/masaar-crm/.git/hooks/post-receive

# 5. Create log file
sudo touch /var/log/masaar-deploy.log
sudo chmod 666 /var/log/masaar-deploy.log
```

---

## ❌ Troubleshooting Deployments

### Deployment Didn't Happen

```bash
# Check if hook exists and is executable
ls -l /opt/masaar-crm/.git/hooks/post-receive

# Check if it's executable
file /opt/masaar-crm/.git/hooks/post-receive
# Should show: ... executable ...

# If not executable, fix it:
chmod +x /opt/masaar-crm/.git/hooks/post-receive
```

### Deployment Failed

```bash
# Check deployment logs
tail -100 /var/log/masaar-deploy.log

# Common issues:
# 1. .env file missing
#    Solution: Create .env on server before pushing

# 2. Docker build failed
#    Check: docker compose logs api

# 3. Port already in use
#    Check: lsof -i :8080

# Manual redeploy if needed:
cd /opt/masaar-crm
docker compose down
docker compose up -d
```

### API Not Responding After Deploy

```bash
# Wait 10-15 seconds (startup takes time)
sleep 15

# Check if running
docker compose ps

# Check logs
docker compose logs api --tail 50

# Restart if needed
docker compose restart api
```

---

## 🔒 Security Considerations

### Only Deploy to 'main' Branch

The hook only deploys when `main` branch is pushed:
```bash
# This triggers deployment
git push origin main

# These do NOT trigger deployment
git push origin develop
git push origin feature/new-feature
```

### Limited Access

Only users with SSH access to the server can deploy via git push.

### Deployment Logs

All deployments are logged:
```bash
# See who deployed what and when
cat /var/log/masaar-deploy.log
```

---

## 📋 Best Practices

### Before Pushing Code

```bash
# 1. Test locally
npm run dev      # Frontend
go run ./cmd/server  # Backend

# 2. Run linter
npm run lint

# 3. Create meaningful commits
git commit -m "feat: Add contact creation via WhatsApp"

# 4. Push to deploy
git push origin main
```

### After Pushing Code

```bash
# 1. Monitor deployment
tail -f /var/log/masaar-deploy.log

# 2. Verify it worked
curl https://api.yourdomain.com/api/v1/stats

# 3. Check frontend
curl https://crm.yourdomain.com
```

### Rollback (if needed)

```bash
# Push previous commit to rollback
git reset --soft HEAD~1
git push -f origin main

# OR, manually restart old version:
cd /opt/masaar-crm
docker compose restart
```

---

## 📊 Advanced: Manual Git Push vs Auto-Deploy

### Option 1: Git Push to Server (Recommended for MVP)

```bash
# Every push auto-deploys
git push origin main

# Pros:
# ✅ Instant deployment
# ✅ No CI/CD setup needed
# ✅ Logs deployment directly on server

# Cons:
# ❌ Any push deploys immediately (no testing)
# ❌ Failed builds affect production
```

### Option 2: GitHub Actions (For Teams)

Add a GitHub Actions workflow that:
- Runs tests
- Builds Docker images
- Pushes to DockerHub
- Notifies your server to pull and restart

Would you like me to create a GitHub Actions workflow?

---

## 🚀 Example Deployment

### Scenario: You Fixed a Bug

```bash
# 1. Make changes
nano internal/api/handler/contact.go

# 2. Test locally
go run ./cmd/server

# 3. Commit and push
git add internal/api/handler/contact.go
git commit -m "fix: Handle empty contact names correctly"
git push origin main

# 4. Server receives push
# 5. Hook runs automatically:
#    - Pulls code
#    - Rebuilds API image
#    - Restarts container
#    - Runs health checks
# 6. Within 50 seconds, new code is live!

# 7. Verify
curl https://api.yourdomain.com/api/v1/contacts
# Your fix is now deployed
```

---

## 📝 Logs Reference

### Successful Deployment Log

```
[2026-04-27 15:30:45] ==========================================
[2026-04-27 15:30:45] Masaar CRM Auto-Deploy Hook Triggered
[2026-04-27 15:30:45] ==========================================
[2026-04-27 15:30:45] Branch pushed: main (revision: a1b2c3d)
[2026-04-27 15:30:45] Pulling latest code...
[2026-04-27 15:30:47] Code updated
[2026-04-27 15:30:47] .env file found
[2026-04-27 15:30:47] Rebuilding Docker images...
[2026-04-27 15:31:05] [SUCCESS] Docker images rebuilt
[2026-04-27 15:31:05] Stopping old containers...
[2026-04-27 15:31:10] Stopping old containers...
[2026-04-27 15:31:12] [SUCCESS] Containers started
[2026-04-27 15:31:22] Running health checks...
[2026-04-27 15:31:22] [SUCCESS] API is responding
[2026-04-27 15:31:22] Final container status:
[2026-04-27 15:31:23] [SUCCESS] DEPLOYMENT COMPLETE
```

### Failed Deployment Log

```
[2026-04-27 15:35:10] [ERROR] .env file missing!
```

---

## 🎯 Summary

| Task | Command |
|------|---------|
| Setup automation | `sudo bash setup-git-automation.sh` |
| Deploy code | `git push origin main` |
| Watch deployment | `tail -f /var/log/masaar-deploy.log` |
| Check status | `docker compose ps` |
| View history | `cat /var/log/masaar-deploy.log` |
| Manual deploy | `cd /opt/masaar-crm && docker compose up -d` |

---

**Git automation is optional but highly recommended for production!**

Set it up once after initial deployment, then you have instant deployments whenever you push code.
