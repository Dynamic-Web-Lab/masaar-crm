# Masaar CRM — Operations Guide

Quick reference for managing Masaar CRM in production.

## Deployment Directory

All commands run from: `/opt/masaar-crm`

```bash
cd /opt/masaar-crm
```

---

## Container Management

### View status
```bash
docker compose ps
```

### View logs (real-time)
```bash
docker compose logs -f              # All services
docker compose logs -f api          # Just API
docker compose logs -f web          # Just Frontend
docker compose logs -f postgres     # Just Database
docker compose logs -f ollama       # Just Ollama
```

### View logs (last 50 lines, no follow)
```bash
docker compose logs --tail 50
docker compose logs --tail 50 api
```

### Start all services
```bash
docker compose up -d
```

### Stop all services
```bash
docker compose down
```

### Restart a service
```bash
docker compose restart api          # Restart API
docker compose restart web          # Restart Frontend
docker compose restart postgres     # Restart Database
```

### Rebuild and restart (after code changes)
```bash
docker compose up --build -d
```

---

## Database Operations

### Connect to PostgreSQL
```bash
docker exec -it masaar-postgres psql -U masaar -d masaar
```

### Useful PostgreSQL commands
```sql
-- List all tables
\dt

-- Show users
SELECT * FROM "user";

-- Show contacts
SELECT id, name, phone FROM contact LIMIT 10;

-- Show leads
SELECT id, title, stage FROM lead LIMIT 10;

-- Check database size
SELECT pg_size_pretty(pg_database_size('masaar'));

-- Exit
\q
```

### Backup database
```bash
docker exec masaar-postgres pg_dump -U masaar masaar > backup-$(date +%Y%m%d-%H%M%S).sql

# Compress backup
gzip backup-*.sql
```

### Restore from backup
```bash
# Stop services
docker compose down

# Restore
gunzip < backup-20260427-120000.sql.gz | \
  docker exec -i masaar-postgres psql -U masaar masaar

# Start services
docker compose up -d
```

---

## Redis Operations

### Connect to Redis
```bash
docker exec -it masaar-redis redis-cli -p 6380 -a PASSWORD
```

### Check Redis status
```bash
docker exec masaar-redis redis-cli -p 6380 PING
```

### Clear cache
```bash
docker exec masaar-redis redis-cli -p 6380 FLUSHDB
```

### Monitor Redis in real-time
```bash
docker exec -it masaar-redis redis-cli -p 6380 MONITOR
```

---

## API Health Checks

### API endpoints to monitor
```bash
# Overall stats
curl https://api.yourdomain.com/api/v1/stats

# Database health
curl https://api.yourdomain.com/api/v1/users/me \
  -H "Authorization: Bearer YOUR_TOKEN"

# Check API docs
https://api.yourdomain.com/docs
```

### Test webhook (WhatsApp)
```bash
# Verify webhook is responding
curl https://api.yourdomain.com/webhooks/whatsapp \
  -H "hub.verify_token: YOUR_VERIFY_TOKEN" \
  -H "hub.challenge: test"
```

---

## Ollama Management

### Check if model is loaded
```bash
curl http://localhost:11434/api/tags
```

### Pull a different model
```bash
docker exec masaar-ollama ollama pull mistral
```

### Check Ollama logs
```bash
docker compose logs ollama
```

### Test Ollama
```bash
curl http://localhost:11434/api/generate -d '{
  "model": "llama3",
  "prompt": "What is Masaar CRM?",
  "stream": false
}'
```

---

## Monitoring & Performance

### Check disk usage
```bash
df -h /

# If > 80% full, clean up:
docker system prune -a
```

### Check memory usage
```bash
free -h
docker stats
```

### Check CPU usage
```bash
top -b -n 1 | head -20
docker stats
```

### Monitor in real-time
```bash
# Docker dashboard
docker stats

# System resources
watch -n 2 'df -h / && echo "---" && free -h'
```

---

## Backup & Disaster Recovery

### Automated daily backups
Create a cron job:

```bash
sudo crontab -e

# Add this line (runs at 2 AM daily):
0 2 * * * /opt/masaar-crm/backup.sh
```

Create `/opt/masaar-crm/backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR="/opt/masaar-crm/backups"
mkdir -p $BACKUP_DIR

# Database backup
docker exec masaar-postgres pg_dump -U masaar masaar | \
  gzip > $BACKUP_DIR/db-$(date +%Y%m%d-%H%M%S).sql.gz

# Keep only last 7 days
find $BACKUP_DIR -name "db-*.sql.gz" -mtime +7 -delete

echo "Backup completed at $(date)" >> $BACKUP_DIR/backup.log
```

### Full system recovery
```bash
# Stop containers
docker compose down

# Restore database
gunzip < backups/db-LATEST.sql.gz | \
  docker exec -i masaar-postgres psql -U masaar masaar

# Start containers
docker compose up -d

# Verify
docker compose ps
```

---

## Troubleshooting

### API won't start
```bash
# Check logs
docker compose logs api --tail 100

# Common issues:
# 1. Database not ready: Check postgres health
docker compose ps postgres

# 2. Port conflict: Check if 8080 is in use
lsof -i :8080

# 3. Environment variables wrong: Check .env file
cat .env | grep DATABASE_URL
```

### Frontend blank/not loading
```bash
# Check API URL is correct
docker compose logs web --tail 50

# Check if API is responding
curl http://localhost:8080/api/v1/stats

# Verify NEXT_PUBLIC_API_URL in .env or docker-compose
```

### Database connection errors
```bash
# Test database
docker exec masaar-postgres psql -U masaar -c "SELECT 1"

# Check password in .env matches database
grep DATABASE_URL .env
grep POSTGRES_PASSWORD docker-compose.prod.override.yml
```

### Ollama taking forever
```bash
# Check progress
docker compose logs ollama -f

# Ollama downloading model can take 10-20 minutes
# Be patient! Don't restart it during download.

# If stuck, restart:
docker compose restart ollama
```

### Port conflicts
```bash
# Find what's using a port
lsof -i :8080      # API
lsof -i :3000      # Frontend
lsof -i :5432      # Database
lsof -i :6380      # Redis

# Kill process (careful!)
kill -9 <PID>
```

### Out of disk space
```bash
# Check usage
df -h /

# Clean up Docker
docker system prune -a --force

# Remove old backups
rm -f /opt/masaar-crm/backups/db-*.sql.gz

# Remove old logs
docker compose logs --tail 0
```

---

## Security

### Change default admin password
1. Login at `https://crm.yourdomain.com`
2. Email: `admin@masaar.local`
3. Password: `changeme`
4. Go to Settings → Change Password

### Rotate secrets
```bash
# Generate new JWT secret
openssl rand -base64 32

# Update .env
nano .env
# Update JWT_SECRET

# Restart API
docker compose restart api
```

### Check for open ports
```bash
# Should only expose 80, 443
sudo ss -tulpn | grep LISTEN

# If other ports exposed, update docker-compose.yml
```

### View security logs
```bash
# Check nginx logs
sudo tail -100 /var/log/nginx/masaar-api-access.log
sudo tail -100 /var/log/nginx/masaar-api-error.log
```

---

## Update & Maintenance

### Update Docker images
```bash
# Pull latest images
docker compose pull

# Rebuild and restart
docker compose up --build -d
```

### Update application code
```bash
# Pull latest from git
git pull origin main

# Rebuild and restart
docker compose up --build -d

# Or manually:
docker compose down
docker compose up --build -d
```

### View version information
```bash
# API version
curl https://api.yourdomain.com/api/v1/stats | jq .

# Docker images
docker compose images

# Git revision
git log --oneline | head -5
```

---

## Emergency Procedures

### Complete reset (WARNING: Deletes all data)
```bash
# Stop services
docker compose down

# Remove volumes (THIS DELETES DATA!)
docker compose down -v

# Clear database directory
rm -rf volumes/postgres/*

# Restart
docker compose up -d
```

### Restart everything
```bash
docker compose restart
```

### Force restart (if hung)
```bash
docker compose down --timeout 10
docker compose up -d
```

### Check service health
```bash
docker compose ps

# All should show "Up" status
# If any show "Exited", check logs:
docker compose logs <service_name>
```

---

## Useful Commands Reference

| Task | Command |
|------|---------|
| View all logs | `docker compose logs --tail 100` |
| Follow API logs | `docker compose logs -f api` |
| Restart API | `docker compose restart api` |
| Restart all | `docker compose restart` |
| Database backup | `docker exec masaar-postgres pg_dump -U masaar masaar \| gzip > backup.sql.gz` |
| Database restore | `gunzip < backup.sql.gz \| docker exec -i masaar-postgres psql -U masaar masaar` |
| Check disk | `df -h /` |
| Check memory | `free -h` |
| Check running services | `docker compose ps` |
| Stop all services | `docker compose down` |
| Start all services | `docker compose up -d` |
| View container stats | `docker stats` |
| Connect to DB | `docker exec -it masaar-postgres psql -U masaar -d masaar` |

---

## Support

For issues not listed here:
1. Check logs: `docker compose logs --tail 100`
2. Check disk/memory: `df -h /`, `free -h`
3. Restart service: `docker compose restart <service>`
4. Check DEPLOYMENT.md or DEPLOYMENT_CHECKLIST.md

---

**Last updated:** 2026-04-27
