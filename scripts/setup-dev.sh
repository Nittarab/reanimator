#!/bin/bash

set -e

PROJECT_NAME="ai-sre-platform"

echo "🔧 Setting up AI SRE Platform development environment..."
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Copying from .env.example..."
    cp .env.example .env
    echo "✅ Created .env file with default values"
    echo ""
fi

# Load environment variables
set -a
source .env 2>/dev/null || true
set +a

# Stop any existing containers
echo "🧹 Cleaning up existing containers..."
docker-compose -p $PROJECT_NAME down 2>/dev/null || true
echo ""

# Start base services (postgres, redis)
echo "📦 Starting database and cache services..."
docker-compose -p $PROJECT_NAME up -d postgres redis
echo ""

# Wait for services to be healthy
echo "⏳ Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
  if docker exec ai-sre-postgres pg_isready -U postgres > /dev/null 2>&1; then
    echo "✅ PostgreSQL is ready"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "❌ PostgreSQL did not become ready in time"
    exit 1
  fi
  sleep 1
done

echo "⏳ Waiting for Redis to be ready..."
for i in {1..30}; do
  if docker exec ai-sre-redis redis-cli ping > /dev/null 2>&1; then
    echo "✅ Redis is ready"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "❌ Redis did not become ready in time"
    exit 1
  fi
  sleep 1
done
echo ""

# Run database migrations
echo "🔄 Running database migrations..."
cd incident-service
DATABASE_HOST=localhost \
DATABASE_PORT=5432 \
DATABASE_NAME=ai_sre \
DATABASE_USER=postgres \
DATABASE_PASSWORD=postgres \
DATABASE_SSL_MODE=disable \
GITHUB_TOKEN=${GITHUB_TOKEN:-dummy_token_for_dev} \
CONFIG_PATH=../config.yaml \
go run cmd/migrate/main.go
cd ..
echo ""

echo "✅ Development environment setup complete!"
echo ""
echo "Infrastructure services are running:"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"
echo ""
echo "Next steps:"
echo "  1. Start the incident service: cd incident-service && go run cmd/server/main.go"
echo "  2. Start the dashboard: cd dashboard && npm run dev"
echo "  3. Or use docker-compose: docker-compose -p $PROJECT_NAME -f docker-compose.yml -f docker-compose.dev.yml up"
echo ""
