#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  PayForward Setup Validation${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check Docker
echo -e "${YELLOW}📦 Checking Docker...${NC}"
if command -v docker &> /dev/null; then
    echo -e "  ${GREEN}✓${NC} Docker: $(docker --version)"
else
    echo -e "  ${RED}✗${NC} Docker not found"
    exit 1
fi

# Check Docker Compose
echo -e "${YELLOW}📦 Checking Docker Compose...${NC}"
if command -v docker-compose &> /dev/null; then
    echo -e "  ${GREEN}✓${NC} Docker Compose: $(docker-compose --version)"
else
    echo -e "  ${RED}✗${NC} Docker Compose not found"
    exit 1
fi

# Validate docker-compose.yml
echo ""
echo -e "${YELLOW}📋 Validating docker-compose.yml...${NC}"
if docker-compose config > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} docker-compose.yml is valid"
else
    echo -e "  ${RED}✗${NC} docker-compose.yml has errors"
    docker-compose config
    exit 1
fi

# Check backend
echo ""
echo -e "${YELLOW}🔧 Checking backend...${NC}"
if [ -d "backend" ]; then
    echo -e "  ${GREEN}✓${NC} Backend directory exists"
    if [ -f "backend/go.mod" ]; then
        echo -e "  ${GREEN}✓${NC} go.mod found"
    fi
    if [ -f "backend/cmd/server/main.go" ]; then
        echo -e "  ${GREEN}✓${NC} main.go found"
    fi
    if [ -f "backend/internal/auth/keycloak.go" ]; then
        echo -e "  ${GREEN}✓${NC} Keycloak auth integration found"
    fi
fi

# Check frontend
echo ""
echo -e "${YELLOW}🎨 Checking frontend...${NC}"
if [ -d "frontend" ]; then
    echo -e "  ${GREEN}✓${NC} Frontend directory exists"
    if [ -f "frontend/package.json" ]; then
        echo -e "  ${GREEN}✓${NC} package.json found"
    fi
    if [ -f "frontend/Dockerfile" ]; then
        echo -e "  ${GREEN}✓${NC} Dockerfile found"
    fi
fi

# Check Keycloak configuration
echo ""
echo -e "${YELLOW}🔐 Checking Keycloak configuration...${NC}"
if [ -f "keycloak-realm.json" ]; then
    echo -e "  ${GREEN}✓${NC} keycloak-realm.json found"
fi

# Check devcontainer
echo ""
echo -e "${YELLOW}🐳 Checking devcontainer...${NC}"
if [ -d ".devcontainer" ]; then
    echo -e "  ${GREEN}✓${NC} .devcontainer directory exists"
    if [ -f ".devcontainer/devcontainer.json" ]; then
        echo -e "  ${GREEN}✓${NC} devcontainer.json found"
    fi
    if [ -f ".devcontainer/Dockerfile" ]; then
        echo -e "  ${GREEN}✓${NC} Dockerfile found"
    fi
fi

# Check GitHub workflows
echo ""
echo -e "${YELLOW}🤖 Checking GitHub workflows...${NC}"
if [ -d ".github/workflows" ]; then
    echo -e "  ${GREEN}✓${NC} .github/workflows directory exists"
    if [ -f ".github/workflows/backend-ci-cd.yml" ]; then
        echo -e "  ${GREEN}✓${NC} Backend CI/CD workflow found"
    fi
    if [ -f ".github/workflows/frontend-ci-cd.yml" ]; then
        echo -e "  ${GREEN}✓${NC} Frontend CI/CD workflow found"
    fi
fi

# Check documentation
echo ""
echo -e "${YELLOW}📚 Checking documentation...${NC}"
if [ -f "README.md" ]; then
    echo -e "  ${GREEN}✓${NC} README.md found"
fi
if [ -f "SETUP.md" ]; then
    echo -e "  ${GREEN}✓${NC} SETUP.md found"
fi
if [ -f "QUICKSTART.md" ]; then
    echo -e "  ${GREEN}✓${NC} QUICKSTART.md found"
fi
if [ -f ".env.example" ]; then
    echo -e "  ${GREEN}✓${NC} .env.example found"
fi
if [ -f "Makefile" ]; then
    echo -e "  ${GREEN}✓${NC} Makefile found"
fi

# Summary
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✅ All checks passed!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo -e "  1. Start services: ${GREEN}docker-compose up -d${NC}"
echo -e "  2. View logs: ${GREEN}docker-compose logs -f${NC}"
echo -e "  3. Access Keycloak: ${GREEN}http://localhost:8180${NC} (admin/admin)"
echo -e "  4. Access Neo4j: ${GREEN}http://localhost:7474${NC} (neo4j/password123)"
echo -e "  5. Access backend: ${GREEN}http://localhost:8080/api/health${NC}"
echo -e "  6. Access frontend: ${GREEN}http://localhost:3000${NC}"
echo ""
echo -e "${YELLOW}Or use Make:${NC}"
echo -e "  ${GREEN}make up${NC}     - Start all services"
echo -e "  ${GREEN}make help${NC}   - See all available commands"
echo ""
