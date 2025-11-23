#!/bin/bash

set -e

echo "🚀 Deploying AI SRE Platform to production..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ .env file not found. Please create one from .env.example"
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

# Validate required environment variables
required_vars=("POSTGRES_PASSWORD" "DATABASE_PASSWORD" "GITHUB_TOKEN")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Required environment variable $var is not set"
        echo "ℹ️  Please set $var in your .env file"
        exit 1
    fi
done

# Build images
echo "🔨 Building production images..."
docker-compose -f docker-compose.prod.yml build

# Stop existing containers
echo "🛑 Stopping existing containers..."
docker-compose -f docker-compose.prod.yml down

# Start services
echo "🎯 Starting production services..."
docker-compose -f docker-compose.prod.yml up -d

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
sleep 10

# Run database migrations
echo "🗄️  Running database migrations..."
docker-compose -f docker-compose.prod.yml exec -T incident-service ./migrate || echo "⚠️  Migration failed or already applied"

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
sleep 15

# Check health
echo "🏥 Checking service health..."
curl -f http://localhost:8080/api/v1/health || echo "⚠️  Incident Service health check failed"
curl -f http://localhost:3001/health || echo "⚠️  Dashboard health check failed"
curl -f http://localhost:3002/api/health || echo "⚠️  Demo App health check failed"

echo ""
echo "✅ Production deployment complete!"
echo ""
echo "📊 Dashboard: http://localhost:3001"
echo "🔧 Incident Service API: http://localhost:8080"
echo "📈 Metrics: http://localhost:9090/metrics"
echo "🎮 Demo App: http://localhost:3002"
echo ""
echo "📝 View logs with: docker-compose -f docker-compose.prod.yml logs -f"
echo "🛑 Stop services with: docker-compose -f docker-compose.prod.yml down"
