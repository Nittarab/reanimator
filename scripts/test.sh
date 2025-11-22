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

# Test Incident Service
echo ""
echo "${YELLOW}📦 Testing Incident Service (Go)...${NC}"
cd incident-service
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
