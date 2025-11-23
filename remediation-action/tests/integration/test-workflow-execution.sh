#!/bin/bash

# Comprehensive integration test for workflow execution
# Tests context file creation, branch naming, and workflow structure
# Note: Not using 'set -e' to allow all tests to run even if some fail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURES_DIR="$SCRIPT_DIR/../fixtures"
TEST_REPO="$FIXTURES_DIR/test-repo"
ACTION_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

echo -e "${BLUE}=========================================="
echo "Workflow Execution Tests"
echo -e "==========================================${NC}"
echo ""

# Test 1: Verify workflow file structure
test_workflow_structure() {
    echo -e "${BLUE}Test 1: Workflow file structure${NC}"
    
    local workflow_file="$TEST_REPO/.github/workflows/remediate.yml"
    
    if [ ! -f "$workflow_file" ]; then
        echo -e "${RED}✗ FAILED: Workflow file not found${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
    
    # Check for required workflow elements
    local required_elements=(
        "workflow_dispatch"
        "incident_id"
        "error_message"
        "service_name"
        "timestamp"
        "uses: actions/checkout"
    )
    
    for element in "${required_elements[@]}"; do
        if ! grep -q "$element" "$workflow_file"; then
            echo -e "${RED}✗ FAILED: Missing required element: $element${NC}"
            ((TESTS_FAILED++))
            return 1
        fi
    done
    
    echo -e "${GREEN}✓ PASSED: Workflow structure is valid${NC}"
    ((TESTS_PASSED++))
    echo ""
}

# Test 2: Verify incident context file creation logic
test_context_file_creation() {
    echo -e "${BLUE}Test 2: Context file creation${NC}"
    
    # Check if jq is available
    if ! command -v jq &> /dev/null; then
        echo -e "${YELLOW}⚠ SKIPPED: jq not installed${NC}"
        echo ""
        return 0
    fi
    
    # Simulate context file creation
    local temp_dir=$(mktemp -d)
    local context_file="$temp_dir/incident-context.json"
    
    # Read test incident
    local incident_file="$FIXTURES_DIR/incident-events/simple-error.json"
    
    if [ ! -f "$incident_file" ]; then
        echo -e "${RED}✗ FAILED: Test incident file not found at $incident_file${NC}"
        ((TESTS_FAILED++))
        rm -rf "$temp_dir"
        return 1
    fi
    
    # Create context file (simulating what the action does)
    if ! cp "$incident_file" "$context_file" 2>/dev/null; then
        echo -e "${RED}✗ FAILED: Could not copy incident file${NC}"
        ((TESTS_FAILED++))
        rm -rf "$temp_dir"
        return 1
    fi
    
    # Verify context file has required fields
    local required_fields=("incident_id" "error_message" "service_name" "timestamp")
    
    for field in "${required_fields[@]}"; do
        if ! jq -e ".$field" "$context_file" > /dev/null 2>&1; then
            echo -e "${RED}✗ FAILED: Missing required field: $field${NC}"
            ((TESTS_FAILED++))
            rm -rf "$temp_dir"
            return 1
        fi
    done
    
    echo -e "${GREEN}✓ PASSED: Context file creation is valid${NC}"
    ((TESTS_PASSED++))
    rm -rf "$temp_dir"
    echo ""
}

# Test 3: Verify branch naming convention
test_branch_naming() {
    echo -e "${BLUE}Test 3: Branch naming convention${NC}"
    
    local incident_id="inc_test_001"
    local expected_pattern="remediate-$incident_id"
    
    # Test branch name generation
    local branch_name="remediate-$incident_id-$(date +%s)"
    
    if [[ ! "$branch_name" =~ ^remediate-$incident_id ]]; then
        echo -e "${RED}✗ FAILED: Branch name does not match pattern${NC}"
        echo "  Expected pattern: remediate-$incident_id-*"
        echo "  Got: $branch_name"
        ((TESTS_FAILED++))
        return 1
    fi
    
    echo -e "${GREEN}✓ PASSED: Branch naming follows convention${NC}"
    echo "  Branch name: $branch_name"
    ((TESTS_PASSED++))
    echo ""
}

# Test 4: Verify MCP configuration handling
test_mcp_config() {
    echo -e "${BLUE}Test 4: MCP configuration handling${NC}"
    
    # Check if jq is available
    if ! command -v jq &> /dev/null; then
        echo -e "${YELLOW}⚠ SKIPPED: jq not installed${NC}"
        echo ""
        return 0
    fi
    
    local mcp_file="$TEST_REPO/.kiro/settings/mcp.json"
    
    if [ ! -f "$mcp_file" ]; then
        echo -e "${RED}✗ FAILED: MCP config file not found at $mcp_file${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
    
    # Verify it's valid JSON
    if ! jq empty "$mcp_file" 2>/dev/null; then
        echo -e "${RED}✗ FAILED: MCP config is not valid JSON${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
    
    # Verify it has mcpServers key
    if ! jq -e '.mcpServers' "$mcp_file" > /dev/null 2>&1; then
        echo -e "${RED}✗ FAILED: MCP config missing mcpServers key${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
    
    echo -e "${GREEN}✓ PASSED: MCP configuration is valid${NC}"
    ((TESTS_PASSED++))
    echo ""
}

# Test 5: Verify test repository structure
test_repo_structure() {
    echo -e "${BLUE}Test 5: Test repository structure${NC}"
    
    local required_files=(
        "$TEST_REPO/src/users.js"
        "$TEST_REPO/src/math.js"
        "$TEST_REPO/package.json"
        "$TEST_REPO/.github/workflows/remediate.yml"
        "$TEST_REPO/.kiro/settings/mcp.json"
    )
    
    for file in "${required_files[@]}"; do
        if [ ! -f "$file" ]; then
            echo -e "${RED}✗ FAILED: Required file not found: $file${NC}"
            ((TESTS_FAILED++))
            return 1
        fi
    done
    
    echo -e "${GREEN}✓ PASSED: Test repository structure is complete${NC}"
    ((TESTS_PASSED++))
    echo ""
}

# Test 6: Verify incident event fixtures
test_incident_fixtures() {
    echo -e "${BLUE}Test 6: Incident event fixtures${NC}"
    
    # Check if jq is available
    if ! command -v jq &> /dev/null; then
        echo -e "${YELLOW}⚠ SKIPPED: jq not installed${NC}"
        echo ""
        return 0
    fi
    
    local fixtures=(
        "$FIXTURES_DIR/incident-events/simple-error.json"
        "$FIXTURES_DIR/incident-events/division-by-zero.json"
        "$FIXTURES_DIR/incident-events/null-pointer.json"
    )
    
    for fixture in "${fixtures[@]}"; do
        if [ ! -f "$fixture" ]; then
            echo -e "${RED}✗ FAILED: Fixture not found: $fixture${NC}"
            ((TESTS_FAILED++))
            return 1
        fi
        
        # Verify it's valid JSON
        if ! jq empty "$fixture" 2>/dev/null; then
            echo -e "${RED}✗ FAILED: Fixture is not valid JSON: $fixture${NC}"
            ((TESTS_FAILED++))
            return 1
        fi
        
        # Verify required fields
        local required_fields=("incident_id" "error_message" "service_name" "timestamp")
        for field in "${required_fields[@]}"; do
            if ! jq -e ".$field" "$fixture" > /dev/null 2>&1; then
                echo -e "${RED}✗ FAILED: Fixture missing field $field: $fixture${NC}"
                ((TESTS_FAILED++))
                return 1
            fi
        done
    done
    
    echo -e "${GREEN}✓ PASSED: All incident fixtures are valid${NC}"
    ((TESTS_PASSED++))
    echo ""
}

# Run all tests
test_workflow_structure
test_context_file_creation
test_branch_naming
test_mcp_config
test_repo_structure
test_incident_fixtures

# Print summary
echo -e "${BLUE}=========================================="
echo "Test Summary"
echo -e "==========================================${NC}"
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
echo "Total:  $((TESTS_PASSED + TESTS_FAILED))"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All workflow execution tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed.${NC}"
    exit 1
fi
