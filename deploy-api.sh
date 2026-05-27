#!/bin/bash
# Masaar CRM — Backend (Go API) Deploy Script
# Cross-compiles Go binary locally for linux/amd64, rsync to server,
# builds minimal Docker image there. Uses --env-file on server so all
# secrets are loaded reliably without shell-expansion issues.
#
# Usage: ./deploy-api.sh

set -e

# ── Config ─────────────────────────────────────────────────────────────────────
SSH_KEY="/Users/maidul/info/contabo/backup-buyorsell24-key"
SERVER="root@109.123.240.239"
SERVER_DIR="/root/projects/masaar-crm"
REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="$REPO_DIR/.env"
CONTAINER="masaar-api"
IMAGE="masaar-crm-api"
NETWORK="buyorsell24_masaar_shared"
BINARY="masaar-linux"

# ── Validate env file ──────────────────────────────────────────────────────────
if [ ! -f "$ENV_FILE" ]; then
  echo "❌ ERROR: $ENV_FILE not found"
  exit 1
fi
echo "📋 Using env file: $ENV_FILE"

# ── Step 1: Cross-compile Go binary for linux/amd64 ───────────────────────────
echo ""
echo "🔨 Cross-compiling Go binary for linux/amd64..."
cd "$REPO_DIR"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -o "$BINARY" ./cmd/server

BINARY_SIZE=$(du -sh "$BINARY" | cut -f1)
echo "✅ Binary built: $BINARY ($BINARY_SIZE)"

# ── Step 2: Sync binary, migrations, and .env to server ───────────────────────
echo ""
echo "📤 Syncing binary, migrations, and .env to server..."
ssh -i "$SSH_KEY" "$SERVER" "mkdir -p $SERVER_DIR/api-deploy/migrations"

echo "   Uploading binary..."
rsync -az --progress -e "ssh -i $SSH_KEY" \
  "$REPO_DIR/$BINARY" "$SERVER:$SERVER_DIR/api-deploy/masaar"

echo "   Syncing migrations/..."
rsync -az --delete -e "ssh -i $SSH_KEY" \
  "$REPO_DIR/migrations/" "$SERVER:$SERVER_DIR/api-deploy/migrations/"

echo "   Uploading .env..."
rsync -az -e "ssh -i $SSH_KEY" \
  "$ENV_FILE" "$SERVER:$SERVER_DIR/.env"

echo "✅ Sync complete"

# ── Step 3: Build image and restart on server ──────────────────────────────────
echo ""
echo "🐳 Building image and restarting on server..."

ssh -i "$SSH_KEY" "$SERVER" bash << 'ENDSSH'
set -e

CONTAINER="masaar-api"
IMAGE="masaar-crm-api"
NETWORK="buyorsell24_masaar_shared"
ENV_FILE="/root/projects/masaar-crm/.env"
DEPLOY_DIR="/root/projects/masaar-crm/api-deploy"

cd "$DEPLOY_DIR"
chmod +x masaar

# Write minimal Dockerfile
cat > Dockerfile.deploy << 'DOCKERFILE'
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget
RUN addgroup -S masaar && adduser -S masaar -G masaar
WORKDIR /app
COPY masaar .
COPY migrations ./migrations
RUN chmod +x /app/masaar
USER masaar
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=40s \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1
ENTRYPOINT ["/app/masaar"]
DOCKERFILE

echo "   Tagging current image as rollback..."
docker tag $IMAGE:latest $IMAGE:rollback 2>/dev/null || true

echo "   Building new image..."
docker build -f Dockerfile.deploy -t $IMAGE:latest . 2>&1

echo "   Stopping old container..."
docker stop $CONTAINER 2>/dev/null || true
docker rm $CONTAINER 2>/dev/null || true

echo "   Starting new container..."
docker run -d \
  --name $CONTAINER \
  --restart unless-stopped \
  --network $NETWORK \
  --env-file "$ENV_FILE" \
  $IMAGE:latest

echo "   Waiting for health check (up to 60s)..."
for i in $(seq 1 30); do
  sleep 2
  if docker exec $CONTAINER wget -qO- http://localhost:8080/health > /dev/null 2>&1; then
    echo "   ✅ API healthy"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "   ❌ API failed health check — rolling back..."
    echo "   Last logs from failed container:"
    docker logs $CONTAINER --tail 20 2>&1 || true
    docker rm -f $CONTAINER 2>/dev/null || true
    docker run -d \
      --name $CONTAINER \
      --restart unless-stopped \
      --network $NETWORK \
      --env-file "$ENV_FILE" \
      $IMAGE:rollback
    echo "   ↩️  Rolled back to previous image."
    exit 1
  fi
done

echo ""
echo "   Container status:"
docker ps --filter name=$CONTAINER --format 'table {{.Names}}\t{{.Status}}'
echo ""
echo "   Last 15 log lines:"
docker logs $CONTAINER --tail 15 2>&1 || true
echo ""
echo "   Cleaning up..."
docker rmi $IMAGE:rollback 2>/dev/null || true
docker image prune -f 2>/dev/null || true
ENDSSH

# Cleanup local binary
rm -f "$REPO_DIR/$BINARY"
echo ""
echo "✅ API deployed!"
echo "🌐 https://masaar.dynamicweblab.com/api/v1/stats"
