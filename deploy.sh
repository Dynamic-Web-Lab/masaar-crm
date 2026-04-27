#!/bin/bash

# Masaar CRM — Production Deployment Script
# Usage: bash deploy.sh

set -e  # Exit on error

COLOR_GREEN='\033[0;32m'
COLOR_YELLOW='\033[1;33m'
COLOR_RED='\033[0;31m'
NC='\033[0m' # No Color

log() {
    echo -e "${COLOR_GREEN}[$(date +'%H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${COLOR_RED}[ERROR]${NC} $1" >&2
    exit 1
}

warn() {
    echo -e "${COLOR_YELLOW}[WARN]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    error "This script must be run as root (use: sudo bash deploy.sh)"
fi

# Check prerequisites
log "Checking prerequisites..."

command -v docker >/dev/null 2>&1 || error "Docker not installed"
command -v git >/dev/null 2>&1 || error "Git not installed"

log "✓ Docker and Git available"

# Step 1: Create deployment directory
log "Step 1/6: Setting up deployment directory..."
DEPLOY_DIR="/opt/masaar-crm"

if [ -d "$DEPLOY_DIR" ]; then
    warn "Directory $DEPLOY_DIR already exists"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    mkdir -p "$DEPLOY_DIR"
    log "Created $DEPLOY_DIR"
fi

cd "$DEPLOY_DIR"

# Step 2: Clone repository
log "Step 2/6: Cloning repository..."
if [ -d ".git" ]; then
    log "Git repo already exists, pulling latest..."
    git pull origin main 2>/dev/null || git pull origin master
else
    git clone https://github.com/maidulcu/masaar-crm.git . || \
        error "Failed to clone repository. Check URL and internet connection."
fi

log "✓ Repository cloned/updated"

# Step 3: Check .env
log "Step 3/6: Checking environment configuration..."

if [ ! -f ".env" ]; then
    warn "No .env file found. Creating from template..."
    cp .env.example .env

    cat << "EOF"

${COLOR_YELLOW}ACTION REQUIRED:${NC}
Edit /opt/masaar-crm/.env with your configuration:

    nano /opt/masaar-crm/.env

Required values:
    - DATABASE_URL (change 'masaar' password)
    - REDIS_URL (set redis://redis:6380)
    - JWT_SECRET (run: openssl rand -base64 32)
    - WA_VERIFY_TOKEN (your webhook token)
    - NEXT_PUBLIC_API_URL (https://api.yourdomain.com)

After editing, run this script again.
EOF
    exit 0
fi

log "✓ .env file found"

# Step 4: Create volume directories
log "Step 4/6: Setting up volume directories..."
mkdir -p volumes/{postgres,redis,ollama}
chmod 755 volumes/*
log "✓ Volume directories ready"

# Step 5: Optional Docker cleanup
log "Step 5/6: Checking disk space..."
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
DISK_AVAIL=$(df / | awk 'NR==2 {print $4}')

log "Disk usage: ${DISK_USAGE}%"

if [ $DISK_USAGE -gt 80 ]; then
    warn "Disk usage is above 80%! Consider cleaning up:"
    warn "  docker system prune -a"
    read -p "Run cleanup now? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log "Pruning Docker system..."
        docker system prune -a --force
        log "✓ Cleanup complete"
    fi
fi

# Step 6: Start services
log "Step 6/6: Starting Masaar CRM..."

log "Building and starting containers (this may take 10-20 minutes on first run)..."
docker compose -f docker-compose.yml -f docker-compose.prod.override.yml up -d

log "Waiting for services to be healthy..."
sleep 10

# Check service status
log "Service Status:"
docker compose ps

# Wait for Ollama model download
if docker compose ps ollama | grep -q "Up"; then
    log "Waiting for Ollama model download (this can take 10-15 minutes)..."
    log "Monitor with: docker compose logs ollama -f"
fi

# Test services
log "Running health checks..."

if curl -s http://localhost:8080/api/v1/stats > /dev/null 2>&1; then
    log "✓ API is responding"
else
    warn "API not yet responding (may still be starting)"
fi

cat << "EOF"

${COLOR_GREEN}================================${NC}
  Masaar CRM Deployment Started!
${COLOR_GREEN}================================${NC}

Next steps:

1. Configure Nginx reverse proxy:
   nano /etc/nginx/sites-available/default
   (See DEPLOYMENT.md Step 6 for config)

2. Setup SSL/TLS:
   sudo certbot certonly --nginx -d api.yourdomain.com -d crm.yourdomain.com

3. Update DNS records:
   A Record: api.yourdomain.com → 109.123.240.239
   A Record: crm.yourdomain.com → 109.123.240.239

4. Monitor logs:
   docker compose logs -f

5. Access the application:
   - Frontend: http://localhost:3000 (via reverse proxy: https://crm.yourdomain.com)
   - API: http://localhost:8080 (via reverse proxy: https://api.yourdomain.com)
   - Default login: admin@masaar.local / changeme

${COLOR_YELLOW}IMPORTANT:${NC} Change the default password immediately!

Full deployment guide: /opt/masaar-crm/DEPLOYMENT.md
Checklist: /opt/masaar-crm/DEPLOYMENT_CHECKLIST.md

EOF

log "Deployment complete! See messages above for next steps."
