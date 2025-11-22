#!/bin/bash

set -e

echo "🚀 Starting AI SRE Platform in development mode..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Copying from .env.example..."
    cp .env.example .env
    echo "📝 Please edit .env with your actual credentials before continuing."
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

# Start base services (postgres, redis)
echo "📦 Starting database and cache services..."
docker-compose up -d postgres redis

# Wait for services to be healthy
echo "⏳ Waiting for services to be ready..."
sleep 5

# Run database migrations
echo "🔄 Running database migrations..."
cd incident-service && go run cmd/migrate/main.go && cd ..

# Start all services
echo "🎯 Starting all services..."
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

echo "✅ Development environment is running!"
echo "📊 Dashboard: http://localhost:3000"
echo "🔧 Incident Service: http://localhost:8080"
echo "🎮 Demo App: http://localhost:3001"
echo "📈 Metrics: http://localhost:9090/metrics"
