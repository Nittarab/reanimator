#!/bin/bash

# Setup test database for running tests

set -e

echo "Starting PostgreSQL container for testing..."

# Start PostgreSQL container
docker run -d \
  --name ai-sre-test-db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=ai_sre_test \
  -p 5434:5432 \
  postgres:15-alpine

echo "Waiting for PostgreSQL to be ready..."
sleep 5

# Wait for PostgreSQL to be ready
until docker exec ai-sre-test-db pg_isready -U postgres > /dev/null 2>&1; do
  echo "Waiting for PostgreSQL..."
  sleep 1
done

echo "PostgreSQL test database is ready!"
echo ""
echo "To run tests, use:"
echo "  export TEST_DATABASE_HOST=localhost"
echo "  export TEST_DATABASE_PORT=5434"
echo "  export TEST_DATABASE_USER=postgres"
echo "  export TEST_DATABASE_PASSWORD=postgres"
echo "  export TEST_DATABASE_NAME=ai_sre_test"
echo "  cd incident-service && go test -v ./..."
echo ""
echo "To stop the test database:"
echo "  docker stop ai-sre-test-db && docker rm ai-sre-test-db"
