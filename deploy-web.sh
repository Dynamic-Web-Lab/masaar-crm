#!/bin/bash
# Masaar CRM — Frontend Deploy Script
# Builds Next.js locally on Mac, rsync to server, builds minimal image there.
# Server docker build is skipped — image assembly on server takes ~10s not 5min.
#
# Usage: ./deploy-web.sh

set -e

# ── Config ─────────────────────────────────────────────────────────────────────
SSH_KEY="/Users/maidul/info/contabo/backup-buyorsell24-key"
SERVER="root@109.123.240.239"
SERVER_DIR="/root/projects/masaar-crm"
WEB_DIR="$(cd "$(dirname "$0")/web" && pwd)"
ENV_FILE="$(cd "$(dirname "$0")" && pwd)/.env"
CONTAINER="masaar-web"
IMAGE="masaar-crm-web"
NETWORK="buyorsell24_masaar_shared"

# ── Load env ───────────────────────────────────────────────────────────────────
echo "📋 Loading environment from $ENV_FILE..."
if [ ! -f "$ENV_FILE" ]; then
  echo "❌ ERROR: $ENV_FILE not found"
  exit 1
fi
export $(grep -v '^#' "$ENV_FILE" | grep -v '^$' | xargs 2>/dev/null) || true
echo "   NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL"

# ── Step 1: Build locally ──────────────────────────────────────────────────────
echo ""
echo "🔨 Building Next.js locally (standalone mode)..."
cd "$WEB_DIR"

# Install deps if needed
if [ ! -d node_modules ]; then
  echo "   Installing node_modules..."
  npm install --prefer-offline
fi

NEXT_PUBLIC_API_URL="$NEXT_PUBLIC_API_URL" \
NEXT_TELEMETRY_DISABLED=1 \
NODE_OPTIONS='--max-old-space-size=4096' \
  npm run build

echo "✅ Local build complete"

# ── Step 2: Sync build output to server ───────────────────────────────────────
echo ""
echo "📤 Syncing build to server..."
ssh -i "$SSH_KEY" "$SERVER" "mkdir -p $SERVER_DIR/web-deploy/.next/standalone $SERVER_DIR/web-deploy/.next/static $SERVER_DIR/web-deploy/public"

echo "   Syncing .next/standalone/..."
rsync -az --delete -e "ssh -i $SSH_KEY" \
  "$WEB_DIR/.next/standalone/" "$SERVER:$SERVER_DIR/web-deploy/.next/standalone/"

echo "   Syncing .next/static/..."
rsync -az --delete -e "ssh -i $SSH_KEY" \
  "$WEB_DIR/.next/static/" "$SERVER:$SERVER_DIR/web-deploy/.next/static/"

echo "   Syncing public/..."
rsync -az --delete -e "ssh -i $SSH_KEY" \
  "$WEB_DIR/public/" "$SERVER:$SERVER_DIR/web-deploy/public/" 2>/dev/null || true

echo "✅ Sync complete"

# ── Step 3: Build image and restart on server ──────────────────────────────────
echo ""
echo "🐳 Building image and restarting on server..."
ssh -i "$SSH_KEY" "$SERVER" bash << ENDSSH
set -e
cd $SERVER_DIR/web-deploy

# Write minimal Dockerfile
cat > Dockerfile.deploy << 'DOCKERFILE'
FROM node:20-alpine
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"
RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 nextjs
COPY --chown=nextjs:nodejs .next/standalone ./
COPY --chown=nextjs:nodejs .next/static ./.next/static
COPY --chown=nextjs:nodejs public ./public
USER nextjs
EXPOSE 3000
CMD ["node", "server.js"]
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
  -e NEXT_PUBLIC_API_URL="$NEXT_PUBLIC_API_URL" \
  $IMAGE:latest

echo "   Waiting for container to respond (up to 60s)..."
for i in \$(seq 1 30); do
  sleep 2
  # Use node (always in image) — accept any HTTP response < 500 as healthy
  # BusyBox wget exits non-zero on 3xx redirects so we avoid it here
  HTTP_CODE=\$(docker exec $CONTAINER node -e \
    "require('http').get('http://127.0.0.1:3000/',r=>{process.stdout.write(String(r.statusCode));process.exit(0)}).on('error',()=>{process.stdout.write('0');process.exit(1)})" \
    2>/dev/null || echo '0')
  if [ "\$HTTP_CODE" != "0" ] && [ "\$HTTP_CODE" -lt 500 ] 2>/dev/null; then
    echo "   ✅ Frontend responding (HTTP \$HTTP_CODE)"
    break
  fi
  if [ \$i -eq 30 ]; then
    echo "   ❌ Frontend failed to respond — rolling back..."
    docker logs $CONTAINER --tail 20 2>&1 || true
    docker rm -f $CONTAINER 2>/dev/null || true
    docker run -d \
      --name $CONTAINER \
      --restart unless-stopped \
      --network $NETWORK \
      -e NEXT_PUBLIC_API_URL="$NEXT_PUBLIC_API_URL" \
      $IMAGE:rollback
    echo "   ↩️  Rolled back to previous image."
    exit 1
  fi
done

echo ""
echo "   Container status:"
docker ps --filter name=$CONTAINER --format 'table {{.Names}}\t{{.Status}}'
echo ""
echo "   Last 10 log lines:"
docker logs $CONTAINER --tail 10 2>&1 || true
echo ""
echo "   Cleaning up old images..."
docker rmi $IMAGE:rollback 2>/dev/null || true
docker image prune -f 2>/dev/null || true
ENDSSH

echo ""
echo "✅ Frontend deployed!"
echo "🌐 https://masaar.dynamicweblab.com"
