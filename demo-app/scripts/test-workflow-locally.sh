#!/bin/bash

# Test Demo Remediation Workflow Locally
# This script helps test the workflow configuration before running it in GitHub Actions

set -e

echo "🔍 Demo Remediation Workflow - Local Test"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if we're in the right directory
if [ ! -f "package.json" ]; then
    echo -e "${RED}❌ Error: Must be run from demo-app directory${NC}"
    exit 1
fi

echo "📋 Checking Prerequisites..."
echo ""

# Check for required files
echo "1. Checking workflow file..."
if [ -f "../.github/workflows/demo-remediate.yml" ]; then
    echo -e "   ${GREEN}✓${NC} Workflow file exists"
else
    echo -e "   ${RED}✗${NC} Workflow file not found"
    exit 1
fi

echo "2. Checking MCP configuration..."
if [ -f ".kiro/settings/mcp.json" ]; then
    echo -e "   ${GREEN}✓${NC} MCP configuration exists"
else
    echo -e "   ${RED}✗${NC} MCP configuration not found"
    exit 1
fi

echo "3. Checking remediation strategies..."
STRATEGIES=("division-by-zero.md" "null-pointer.md" "array-processing.md")
for strategy in "${STRATEGIES[@]}"; do
    if [ -f ".kiro/specs/demo-fixes/$strategy" ]; then
        echo -e "   ${GREEN}✓${NC} $strategy exists"
    else
        echo -e "   ${RED}✗${NC} $strategy not found"
        exit 1
    fi
done

echo ""
echo "🔐 Checking Environment Variables..."
echo ""

# Check for Sentry credentials
MISSING_VARS=()

if [ -z "$SENTRY_AUTH_TOKEN" ]; then
    MISSING_VARS+=("SENTRY_AUTH_TOKEN")
fi

if [ -z "$SENTRY_ORG" ]; then
    MISSING_VARS+=("SENTRY_ORG")
fi

if [ -z "$SENTRY_PROJECT" ]; then
    MISSING_VARS+=("SENTRY_PROJECT")
fi

if [ ${#MISSING_VARS[@]} -eq 0 ]; then
    echo -e "${GREEN}✓${NC} All Sentry environment variables are set"
    echo "   - SENTRY_AUTH_TOKEN: ${SENTRY_AUTH_TOKEN:0:10}..."
    echo "   - SENTRY_ORG: $SENTRY_ORG"
    echo "   - SENTRY_PROJECT: $SENTRY_PROJECT"
else
    echo -e "${YELLOW}⚠${NC}  Missing environment variables:"
    for var in "${MISSING_VARS[@]}"; do
        echo "   - $var"
    done
    echo ""
    echo "To set them, run:"
    echo "  export SENTRY_AUTH_TOKEN=your-token"
    echo "  export SENTRY_ORG=your-org"
    echo "  export SENTRY_PROJECT=your-project"
    echo ""
    echo "Or create a .env file in demo-app directory:"
    echo "  SENTRY_AUTH_TOKEN=your-token"
    echo "  SENTRY_ORG=your-org"
    echo "  SENTRY_PROJECT=your-project"
fi

echo ""
echo "🧪 Testing MCP Configuration..."
echo ""

# Test MCP configuration parsing
if command -v jq &> /dev/null; then
    echo "Parsing MCP configuration..."
    if jq empty .kiro/settings/mcp.json 2>/dev/null; then
        echo -e "${GREEN}✓${NC} MCP configuration is valid JSON"
        
        # Check for Sentry server
        if jq -e '.mcpServers.sentry' .kiro/settings/mcp.json > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC} Sentry MCP server is configured"
            
            # Show configuration
            echo ""
            echo "Sentry MCP Server Configuration:"
            jq '.mcpServers.sentry' .kiro/settings/mcp.json
        else
            echo -e "${RED}✗${NC} Sentry MCP server not found in configuration"
        fi
    else
        echo -e "${RED}✗${NC} MCP configuration is invalid JSON"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠${NC}  jq not installed, skipping JSON validation"
    echo "   Install jq to enable validation: brew install jq (macOS) or apt-get install jq (Linux)"
fi

echo ""
echo "📝 Testing Workflow Configuration..."
echo ""

# Check workflow has required environment variables
if grep -q "SENTRY_AUTH_TOKEN" ../.github/workflows/demo-remediate.yml; then
    echo -e "${GREEN}✓${NC} Workflow includes SENTRY_AUTH_TOKEN"
else
    echo -e "${RED}✗${NC} Workflow missing SENTRY_AUTH_TOKEN"
fi

if grep -q "SENTRY_ORG" ../.github/workflows/demo-remediate.yml; then
    echo -e "${GREEN}✓${NC} Workflow includes SENTRY_ORG"
else
    echo -e "${RED}✗${NC} Workflow missing SENTRY_ORG"
fi

if grep -q "SENTRY_PROJECT" ../.github/workflows/demo-remediate.yml; then
    echo -e "${GREEN}✓${NC} Workflow includes SENTRY_PROJECT"
else
    echo -e "${RED}✗${NC} Workflow missing SENTRY_PROJECT"
fi

echo ""
echo "🎯 Test Scenarios"
echo ""
echo "You can test the workflow with these incident scenarios:"
echo ""

echo "1. Division by Zero:"
echo "   Incident ID: test-division-001"
echo "   Error: Invalid average calculation: Infinity"
echo "   Stack trace: at calculateAverage (demo-app/src/routes/buggy.js:15:20)"
echo ""

echo "2. Null Pointer:"
echo "   Incident ID: test-null-pointer-001"
echo "   Error: TypeError: Cannot read property 'name' of undefined"
echo "   Stack trace: at getUserName (demo-app/src/routes/buggy.js:42:15)"
echo ""

echo "3. Array Processing:"
echo "   Incident ID: test-array-001"
echo "   Error: TypeError: Cannot read property 'productId' of undefined"
echo "   Stack trace: at processOrders (demo-app/src/routes/buggy.js:78:25)"
echo ""

echo "📚 Next Steps"
echo ""
echo "To test the workflow in GitHub Actions:"
echo "1. Ensure GitHub Secrets are set:"
echo "   - Go to: Settings → Secrets and variables → Actions"
echo "   - Add: SENTRY_AUTH_TOKEN, SENTRY_ORG, SENTRY_PROJECT"
echo ""
echo "2. Trigger the workflow:"
echo "   - Go to: Actions → Demo Remediation Workflow → Run workflow"
echo "   - Fill in incident details from one of the test scenarios above"
echo "   - Click 'Run workflow'"
echo ""
echo "3. Monitor execution:"
echo "   - Watch the workflow logs"
echo "   - Check for created PR"
echo "   - Review the post-mortem and fix"
echo ""

echo "For detailed instructions, see:"
echo "  - demo-app/.github/DEMO-WORKFLOW-SETUP.md (full setup guide)"
echo "  - demo-app/.github/QUICK-TEST-GUIDE.md (quick reference)"
echo ""

echo -e "${GREEN}✅ Local test complete!${NC}"
