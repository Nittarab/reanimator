#!/bin/bash

# Seed data script for AI SRE Platform
# Populates the database with sample incidents for testing

set -e

API_URL="${API_URL:-http://localhost:8080}"

echo "Seeding database with sample incidents..."

# Function to create an incident via webhook
create_incident() {
  local provider=$1
  local service=$2
  local error_msg=$3
  local severity=$4
  local stack_trace=$5
  
  echo "Creating $severity incident for $service via $provider..."
  
  case $provider in
    "datadog")
      curl -s -X POST "$API_URL/api/v1/webhooks/incidents?provider=datadog" \
        -H "Content-Type: application/json" \
        -d "{
          \"id\": \"$(uuidgen)\",
          \"title\": \"$error_msg\",
          \"body\": \"$error_msg\",
          \"alert_type\": \"error\",
          \"priority\": \"$severity\",
          \"tags\": [\"service:$service\", \"env:production\"],
          \"date_happened\": $(date +%s),
          \"aggregation_key\": \"$service-errors\",
          \"source_type_name\": \"ALERT\"
        }" > /dev/null
      ;;
    "pagerduty")
      curl -s -X POST "$API_URL/api/v1/webhooks/incidents?provider=pagerduty" \
        -H "Content-Type: application/json" \
        -d "{
          \"event\": {
            \"id\": \"$(uuidgen)\",
            \"event_type\": \"incident.triggered\",
            \"resource_type\": \"incident\",
            \"occurred_at\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
            \"data\": {
              \"id\": \"P$(uuidgen | cut -c1-6)\",
              \"type\": \"incident\",
              \"title\": \"$error_msg\",
              \"service\": {
                \"id\": \"PSVC123\",
                \"summary\": \"$service\"
              },
              \"urgency\": \"$severity\",
              \"body\": {
                \"details\": \"$stack_trace\"
              }
            }
          }
        }" > /dev/null
      ;;
    "sentry")
      curl -s -X POST "$API_URL/api/v1/webhooks/incidents?provider=sentry" \
        -H "Content-Type: application/json" \
        -d "{
          \"action\": \"created\",
          \"data\": {
            \"issue\": {
              \"id\": \"$(shuf -i 100000000-999999999 -n 1)\",
              \"title\": \"$error_msg\",
              \"culprit\": \"app/services/$service.js in handler\",
              \"level\": \"error\",
              \"platform\": \"javascript\",
              \"project\": \"$service\"
            },
            \"event\": {
              \"event_id\": \"$(uuidgen)\",
              \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
              \"exception\": {
                \"values\": [
                  {
                    \"type\": \"Error\",
                    \"value\": \"$error_msg\",
                    \"stacktrace\": {
                      \"frames\": [
                        {
                          \"filename\": \"app/services/$service.js\",
                          \"function\": \"handler\",
                          \"lineno\": 42
                        }
                      ]
                    }
                  }
                ]
              },
              \"tags\": [
                [\"environment\", \"production\"],
                [\"service\", \"$service\"]
              ]
            }
          }
        }" > /dev/null
      ;;
  esac
  
  echo "✓ Created incident for $service"
}

# Wait for API to be ready
echo "Waiting for API to be ready..."
for i in {1..30}; do
  if curl -s "$API_URL/api/v1/health" > /dev/null 2>&1; then
    echo "✓ API is ready"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "✗ API did not become ready in time"
    exit 1
  fi
  sleep 1
done

# Create sample incidents with various statuses
# Note: Status transitions happen through workflow webhooks, so we'll create incidents
# and some will be in pending state

# Pending incidents
create_incident "datadog" "api-gateway" "High error rate: NullPointerException in UserService.getUser()" "normal" "at UserService.getUser(UserService.java:45)"
create_incident "sentry" "payment-service" "TypeError: Cannot read property 'amount' of undefined" "error" "at processPayment (payment.js:23)"
create_incident "pagerduty" "auth-service" "Database connection timeout" "high" "at DatabasePool.getConnection(pool.js:89)"

# Create more incidents for different services
create_incident "datadog" "notification-service" "RabbitMQ connection lost" "normal" "at MessageQueue.connect(queue.js:12)"
create_incident "sentry" "user-service" "Validation error: Invalid email format" "error" "at validateEmail (validator.js:56)"
create_incident "pagerduty" "order-service" "Redis cache miss rate exceeded threshold" "low" ""

echo ""
echo "✓ Seed data created successfully!"
echo ""
echo "Created 6 sample incidents across different services and providers"
echo "View them at: http://localhost:3001"
echo ""
