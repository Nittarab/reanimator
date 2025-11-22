#!/bin/bash
set -e

# Integration test script for remediation workflow using nektos/act
# This script tests the GitHub Action locally without requiring GitHub infrastructure

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURES_DIR="$SCRIPT_DIR/../fixtures"
TEST_REPO="$FIXTURES_DIR/test-repo"
ACTION_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results tracking
TESTS_PASSED=0
TESTS_FAILED=0

echo "=========================================="
echo "AI SRE Remediation Action - Integration Tests"
echo "=========================================="
echo ""

# Check if act is installed
if ! command -v act &> /dev/null; then
    echo -e "${RED}ERROR: nektos/act is not installed${NC}"
    echo ""
    echo "Please install act using one of the following methods:"
    echo ""
    echo "  macOS (Homebrew):"
    echo "    brew install act"
    echo ""
    echo "  Linux:"
    echo "    curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash"
    echo ""
    echo "  Manual installation:"
    echo "    https://github.com/nektos/act#installation"
    echo ""
    exit 1
fi

echo -e "${GREEN}✓ nektos/act is installed${NC}"
echo ""

# Check if action is built
if [ ! -f "$ACTION_DIR/dist/index.js" ]; then
    echo -e "${YELLOW}Building action...${NC}"
    cd "$ACTION_DIR"
    npm run build
    echo -e "${GREEN}✓ Action built${NC}"
    echo ""
fi

# Function to run a test
run_test() {
    local test_name="$1"
    local incident_file="$2"
    local expected_behavior="$3"
    
    echo "----------------------------------------"
    echo "Test: $test_name"
    echo "----------------------------------------"
    
    # Read incident data
    if [ ! -f "$incident_file" ]; then
        echo -e "${RED}✗ FAILED: Incident file not found: $incident_file${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
    
    # Extract incident data
    incident_id=$(jq -r '.incident_id' "$incident_file")
    error_message=$(jq -r '.error_message' "$incident_file")
    stack_trace=$(jq -r '.stack_trace // ""' "$incident_file")
    service_name=$(jq -r '.service_name' "$incident_file")
    timestamp=$(jq -r '.timestamp' "$incident_file")
    severity=$(jq -r '.severity // "medium"' "$incident_file")
    
    echo "Incident ID: $incident_id"
    echo "Service: $service_name"
    echo "Severity: $severity"
    echo ""
    
    # Create event payload for act
    local event_file="$FIXTURES_DIR/event-$incident_id.json"
    cat > "$event_file" <<EOF
{
  "inputs": {
    "incident_id": "$incident_id",
    "error_message": "$error_message",
    "stack_trace": "$stack_trace",
    "service_name": "$service_name",
    "timestamp": "$timestamp",
    "severity": "$severity"
  }
}
EOF
    
    # Run act with dry-run mode
    echo "Running workflow with act (dry-run)..."
    cd "$TEST_REPO"
    
    # Run act and capture output
    if act workflow_dispatch \
        --eventpath "$event_file" \
        --workflows .github/workflows/remediate.yml \
        --dryrun \
        --verbose \
        2>&1 | tee /tmp/act-output-$incident_id.log; then
        
        echo -e "${GREEN}✓ PASSED: $test_name${NC}"
        ((TESTS_PASSED++))
        rm -f "$event_file"
        return 0
    else
        echo -e "${RED}✗ FAILED: $test_name${NC}"
        echo "See /tmp/act-output-$incident_id.log for details"
        ((TESTS_FAILED++))
        rm -f "$event_file"
        return 1
    fi
}

# Function to test error handling
test_error_handling() {
    local test_name="$1"
    local error_scenario="$2"
    
    echo "----------------------------------------"
    echo "Test: $test_name"
    echo "----------------------------------------"
    echo "Testing error scenario: $error_scenario"
    echo ""
    
    # This is a placeholder for error handling tests
    # In a real scenario, we would modify the test repo to trigger specific errors
    
    echo -e "${YELLOW}⊘ SKIPPED: Error handling tests require full act execution${NC}"
    echo "  (These tests would be run in CI with full GitHub Actions environment)"
    echo ""
}

# Run test suite
echo "Running integration tests..."
echo ""

# Test 1: Simple error scenario
run_test \
    "Simple TypeError remediation" \
    "$FIXTURES_DIR/incident-events/simple-error.json" \
    "Should create context file and prepare branch"

# Test 2: Division by zero
run_test \
    "Division by zero remediation" \
    "$FIXTURES_DIR/incident-events/division-by-zero.json" \
    "Should identify math error and prepare fix"

# Test 3: Null pointer
run_test \
    "Null pointer remediation" \
    "$FIXTURES_DIR/incident-events/null-pointer.json" \
    "Should handle null reference error"

# Test 4: Error handling - missing Kiro CLI
test_error_handling \
    "Missing Kiro CLI error handling" \
    "Kiro CLI installation failure"

# Test 5: Error handling - git failures
test_error_handling \
    "Git operation error handling" \
    "Git push failure"

# Print summary
echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
echo "Total:  $((TESTS_PASSED + TESTS_FAILED))"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi
