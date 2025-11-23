#!/bin/bash

set -e

PROJECT_NAME="ai-sre-platform"

echo "🚀 Starting AI SRE Platform in development mode..."
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
docker-compose -f docker-compose.dev.yml down 2>/dev/null || true
echo ""

# Start base services (postgres, redis)
echo "📦 Starting database and cache services..."
docker-compose -f docker-compose.dev.yml up -d postgres redis
echo ""

# Wait for services to be healthy
echo "⏳ Waiting for services to be ready..."
for i in {1..30}; do
  if docker exec ai-sre-postgres-dev pg_isready -U postgres > /dev/null 2>&1; then
    echo "✅ PostgreSQL is ready"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "❌ PostgreSQL did not become ready in time"
    exit 1
  fi
  sleep 1
done

for i in {1..30}; do
  if docker exec ai-sre-redis-dev redis-cli ping > /dev/null 2>&1; then
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
docker exec ai-sre-postgres-dev psql -U postgres -d ai_sre -c "SELECT 1" > /dev/null 2>&1 || {
  echo "Database ai_sre already exists"
}

# Run migrations from the host
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
echo "✅ Migrations completed"
echo ""

# Start all services
echo "🎯 Starting all services..."
echo "   This will start:"
echo "   - Incident Service (Go backend)"
echo "   - Dashboard (React frontend)"
echo "   - Demo App (Node.js)"
echo ""

# Start services in background
docker-compose -f docker-compose.dev.yml up -d

# Wait for incident service to be ready
echo "⏳ Waiting for Incident Service to be ready..."
for i in {1..30}; do
  if curl -s http://localhost:8080/api/v1/health > /dev/null 2>&1; then
    echo "✅ Incident Service is ready"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "❌ Incident Service did not become ready in time"
    docker-compose -p $PROJECT_NAME logs incident-service
    exit 1
  fi
  sleep 1
done
echo ""

# Seed the database with sample data
echo "🌱 Seeding database with sample incidents..."
./scripts/seed-data.sh
echo ""

echo "✅ All services are running!"
echo ""
echo "   Dashboard:        http://localhost:3001"
echo "   Incident Service: http://localhost:8080"
echo "   Demo App:         http://localhost:3002"
echo ""
echo "Press Ctrl+C to stop all services"
echo ""

# Follow logs
docker-compose -f docker-compose.dev.yml logs -f

echo ""
echo "✅ Development environment stopped"
