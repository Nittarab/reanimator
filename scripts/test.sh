#!/bin/bash

set -e

echo "🧪 Running all tests for AI SRE Platform..."

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track test results
FAILED=0

# Cleanup function
cleanup() {
    echo ""
    echo "${YELLOW}🧹 Cleaning up test containers...${NC}"
    docker stop ai-sre-test-postgres 2>/dev/null || true
    docker rm ai-sre-test-postgres 2>/dev/null || true
    docker stop ai-sre-test-redis 2>/dev/null || true
    docker rm ai-sre-test-redis 2>/dev/null || true
}

# Register cleanup on exit
trap cleanup EXIT

# Start test database containers
echo ""
echo "${YELLOW}🐳 Starting test database containers...${NC}"

# Stop any existing test containers
cleanup

# Start PostgreSQL test container
docker run -d \
  --name ai-sre-test-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=ai_sre_test \
  -p 5434:5432 \
  postgres:15-alpine > /dev/null

# Start Redis test container
docker run -d \
  --name ai-sre-test-redis \
  -p 6380:6379 \
  redis:7-alpine > /dev/null

# Wait for PostgreSQL to be ready
echo "${YELLOW}⏳ Waiting for PostgreSQL to be ready...${NC}"
for i in {1..30}; do
  if docker exec ai-sre-test-postgres pg_isready -U postgres > /dev/null 2>&1; then
    echo "${GREEN}✅ PostgreSQL is ready${NC}"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "${RED}❌ PostgreSQL did not become ready in time${NC}"
    exit 1
  fi
  sleep 1
done

# Wait for Redis to be ready
echo "${YELLOW}⏳ Waiting for Redis to be ready...${NC}"
for i in {1..30}; do
  if docker exec ai-sre-test-redis redis-cli ping > /dev/null 2>&1; then
    echo "${GREEN}✅ Redis is ready${NC}"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "${RED}❌ Redis did not become ready in time${NC}"
    exit 1
  fi
  sleep 1
done

# Run database migrations for test database
echo ""
echo "${YELLOW}🔄 Running database migrations...${NC}"
cd incident-service
export TEST_DATABASE_HOST=localhost
export TEST_DATABASE_PORT=5434
export TEST_DATABASE_NAME=ai_sre_test
export TEST_DATABASE_USER=postgres
export TEST_DATABASE_PASSWORD=postgres
export TEST_DATABASE_SSL_MODE=disable
export GITHUB_TOKEN=dummy_token_for_test
export CONFIG_PATH=../config.yaml
go run cmd/migrate/main.go
cd ..
echo "${GREEN}✅ Migrations completed${NC}"

# Test Incident Service
echo ""
echo "${YELLOW}📦 Testing Incident Service (Go)...${NC}"
cd incident-service
export TEST_DATABASE_HOST=localhost
export TEST_DATABASE_PORT=5434
export TEST_DATABASE_NAME=ai_sre_test
export TEST_DATABASE_USER=postgres
export TEST_DATABASE_PASSWORD=postgres
export TEST_DATABASE_SSL_MODE=disable
export TEST_REDIS_HOST=localhost
export TEST_REDIS_PORT=6380
if go test -v -race -coverprofile=coverage.out ./...; then
    echo "${GREEN}✅ Incident Service tests passed${NC}"
    go tool cover -func=coverage.out | grep total
else
    echo "${RED}❌ Incident Service tests failed${NC}"
    FAILED=1
fi
cd ..

# Test Dashboard
echo ""
echo "${YELLOW}🎨 Testing Dashboard (React)...${NC}"
cd dashboard
if [ -d "node_modules" ]; then
    if npm test -- --run; then
        echo "${GREEN}✅ Dashboard tests passed${NC}"
    else
        echo "${RED}❌ Dashboard tests failed${NC}"
        FAILED=1
    fi
else
    echo "${YELLOW}⚠️  Dashboard dependencies not installed. Run 'npm install' first.${NC}"
fi
cd ..

# Test Demo App
echo ""
echo "${YELLOW}🎮 Testing Demo App (Node.js)...${NC}"
cd demo-app
if [ -d "node_modules" ]; then
    if npm test -- --run; then
        echo "${GREEN}✅ Demo App tests passed${NC}"
    else
        echo "${RED}❌ Demo App tests failed${NC}"
        FAILED=1
    fi
else
    echo "${YELLOW}⚠️  Demo App dependencies not installed. Run 'npm install' first.${NC}"
fi
cd ..

# Test Remediation Action
echo ""
echo "${YELLOW}⚙️  Testing Remediation Action (TypeScript)...${NC}"
cd remediation-action
if [ -d "node_modules" ]; then
    if npm test; then
        echo "${GREEN}✅ Remediation Action tests passed${NC}"
    else
        echo "${RED}❌ Remediation Action tests failed${NC}"
        FAILED=1
    fi
else
    echo "${YELLOW}⚠️  Remediation Action dependencies not installed. Run 'npm install' first.${NC}"
fi
cd ..

# Summary
echo ""
echo "================================"
if [ $FAILED -eq 0 ]; then
    echo "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo "${RED}❌ Some tests failed${NC}"
    exit 1
fi
