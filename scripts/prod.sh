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
required_vars=("DATABASE_URL" "REDIS_URL" "GITHUB_TOKEN" "ENCRYPTION_KEY")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Required environment variable $var is not set"
        exit 1
    fi
done

# Build images
echo "🔨 Building production images..."
docker-compose -f docker-compose.yml -f docker-compose.prod.yml build

# Stop existing containers
echo "🛑 Stopping existing containers..."
docker-compose -f docker-compose.yml -f docker-compose.prod.yml down

# Start services
echo "🎯 Starting production services..."
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
sleep 10

# Check health
echo "🏥 Checking service health..."
curl -f http://localhost:8080/api/v1/health || echo "⚠️  Incident Service health check failed"
curl -f http://localhost:3000 || echo "⚠️  Dashboard health check failed"
curl -f http://localhost:3001/health || echo "⚠️  Demo App health check failed"

echo "✅ Production deployment complete!"
echo "📊 Dashboard: http://localhost:3000"
echo "🔧 Incident Service: http://localhost:8080"
echo "🎮 Demo App: http://localhost:3001"
echo ""
echo "📝 View logs with: docker-compose -f docker-compose.yml -f docker-compose.prod.yml logs -f"
