#!/bin/bash

# Setup Git Automation for Masaar CRM
# Run this ONCE on the production server after initial deployment
#
# Usage: bash setup-git-automation.sh
#
# What it does:
# 1. Sets up .git/hooks/post-receive for auto-deployment on push
# 2. Creates deployment log file
# 3. Tests git automation
#
# After running this:
# - Every `git push` to main branch triggers auto-deploy
# - Deployment logs stored in /var/log/masaar-deploy.log
# - Monitor with: tail -f /var/log/masaar-deploy.log

set -e

REPO_DIR="/opt/masaar-crm"
HOOKS_DIR="$REPO_DIR/.git/hooks"
POST_RECEIVE_HOOK="$HOOKS_DIR/post-receive"
DEPLOY_LOG="/var/log/masaar-deploy.log"

COLOR_GREEN='\033[0;32m'
COLOR_YELLOW='\033[1;33m'
COLOR_RED='\033[0;31m'
NC='\033[0m'

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
    error "This script must be run as root (use: sudo bash setup-git-automation.sh)"
fi

# Check if repo exists
if [ ! -d "$REPO_DIR/.git" ]; then
    error "Git repository not found at $REPO_DIR. Run initial deployment first."
fi

log "Setting up Git Automation for Masaar CRM..."
log "Repository: $REPO_DIR"

# Step 1: Create hooks directory if it doesn't exist
if [ ! -d "$HOOKS_DIR" ]; then
    mkdir -p "$HOOKS_DIR"
    log "Created .git/hooks directory"
fi

# Step 2: Copy post-receive hook
log "Installing post-receive hook..."
if [ ! -f "$POST_RECEIVE_HOOK" ]; then
    # If hook doesn't exist, create it from this file
    cat > "$POST_RECEIVE_HOOK" << 'HOOK_EOF'
#!/bin/bash

set -e

REPO_DIR="/opt/masaar-crm"
DEPLOY_LOG="/var/log/masaar-deploy.log"
BRANCH="main"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$DEPLOY_LOG"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$DEPLOY_LOG"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$DEPLOY_LOG"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$DEPLOY_LOG"
}

log "=========================================="
log "Masaar CRM Auto-Deploy Hook Triggered"
log "=========================================="

cd "$REPO_DIR" || error "Cannot cd to $REPO_DIR"

while read oldrev newrev ref; do
    PUSHED_BRANCH=${ref#refs/heads/}

    if [ "$PUSHED_BRANCH" != "$BRANCH" ]; then
        warn "Skipping $PUSHED_BRANCH (only deploying $BRANCH)"
        continue
    fi

    log "Branch pushed: $PUSHED_BRANCH (revision: ${newrev:0:7})"

    log "Pulling latest code..."
    git fetch origin "$BRANCH" || error "Failed to fetch"
    git reset --hard "origin/$BRANCH" || error "Failed to reset"
    success "Code updated"

    if [ ! -f ".env" ]; then
        error ".env file missing!"
    fi
    log ".env file found"

    log "Rebuilding Docker images..."
    docker compose build --no-cache api web 2>&1 | tee -a "$DEPLOY_LOG" || \
        error "Docker build failed"
    success "Docker images rebuilt"

    log "Stopping old containers..."
    docker compose down || warn "No containers to stop"

    log "Starting new containers..."
    docker compose up -d || error "Failed to start containers"
    success "Containers started"

    log "Waiting for services to be healthy..."
    sleep 10

    log "Running health checks..."
    if curl -s http://localhost:8080/api/v1/stats > /dev/null 2>&1; then
        success "API is responding"
    else
        warn "API not yet responding (may still be starting)"
    fi

    log "Final container status:"
    docker compose ps | tee -a "$DEPLOY_LOG"

    log "=========================================="
    success "DEPLOYMENT COMPLETE"
    log "=========================================="

done

exit 0
HOOK_EOF

    chmod +x "$POST_RECEIVE_HOOK"
    log "Post-receive hook installed and made executable"
else
    log "Post-receive hook already exists"
fi

# Step 3: Create deployment log file
if [ ! -f "$DEPLOY_LOG" ]; then
    touch "$DEPLOY_LOG"
    chmod 666 "$DEPLOY_LOG"
    log "Created deployment log at $DEPLOY_LOG"
else
    log "Deployment log already exists at $DEPLOY_LOG"
fi

# Step 4: Verify hook permissions
if [ ! -x "$POST_RECEIVE_HOOK" ]; then
    chmod +x "$POST_RECEIVE_HOOK"
    log "Fixed post-receive hook permissions"
fi

# Step 5: Test git configuration
log "Verifying git configuration..."
cd "$REPO_DIR"

if git remote -v | grep -q origin; then
    log "Git remote 'origin' configured"
else
    error "Git remote 'origin' not configured"
fi

log "Current branch: $(git branch --show-current)"
log "Latest commit: $(git log -1 --oneline)"

# Step 6: Show instructions
cat << EOF

${COLOR_GREEN}════════════════════════════════════════════════${NC}
  Git Automation Setup Complete! ✅
${COLOR_GREEN}════════════════════════════════════════════════${NC}

Post-receive hook installed at:
  $POST_RECEIVE_HOOK

Deployment log location:
  $DEPLOY_LOG

${COLOR_YELLOW}HOW IT WORKS:${NC}
1. Developer pushes code: git push origin main
2. Server receives push in $REPO_DIR
3. post-receive hook triggers automatically
4. Code is pulled, Docker images rebuilt, containers restarted
5. Deployment logged to $DEPLOY_LOG

${COLOR_YELLOW}TO MONITOR DEPLOYMENTS:${NC}
  # Watch logs in real-time during deployment
  tail -f $DEPLOY_LOG

  # View deployment history
  cat $DEPLOY_LOG

  # Check container status
  docker compose ps

${COLOR_YELLOW}TO DEPLOY:${NC}
  # On your local machine:
  git push origin main

  # Server will auto-deploy within seconds
  # Monitor with: tail -f /var/log/masaar-deploy.log

${COLOR_YELLOW}TROUBLESHOOTING:${NC}
  If deployment fails:
  1. Check logs: cat $DEPLOY_LOG
  2. Manual check: docker compose logs --tail 50
  3. Manual redeploy: cd $REPO_DIR && docker compose down && docker compose up -d

${COLOR_GREEN}════════════════════════════════════════════════${NC}

EOF

success "Git automation is now active!"
success "Every push to 'main' branch will auto-deploy"
