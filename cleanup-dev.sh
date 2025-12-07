#!/bin/bash

echo "🧹 Starting complete cleanup..."

# 1. Stop all containers
echo "Stopping containers..."
docker compose -f docker-compose.dev.yml down -v

# 2. Remove tmp directories on host
echo "Removing tmp directories..."
rm -rf tmp/
rm -rf .air_tmp/
mkdir -p tmp

# 3. Remove vendor if exists
echo "Removing vendor..."
rm -rf vendor/

# 4. Remove all auth images
echo "Removing auth images..."
docker images | grep auth | awk '{print $3}' | xargs -r docker rmi -f

# 5. Remove all volumes
echo "Removing volumes..."
docker volume prune -f

# 6. Remove all build cache
echo "Removing build cache..."
docker builder prune -a -f

# 7. Remove Go cache from any existing containers
echo "Clearing Go cache..."
docker compose -f docker-compose.dev.yml run --rm api go clean -cache -modcache -testcache 2>/dev/null || true

# 8. Rebuild without cache
echo "Rebuilding without cache..."
docker compose -f docker-compose.dev.yml build --no-cache

# 9. Start fresh
echo "Starting fresh containers..."
docker compose -f docker-compose.dev.yml up

echo "✅ Complete cleanup done!"