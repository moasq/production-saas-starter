#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Store the root directory
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$ROOT_DIR/go-b2b-starter"

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}🛑 Shutting down dependency services...${NC}"
    echo -e "${BLUE}→ Stopping Docker containers...${NC}"
    cd "$BACKEND_DIR"
    make stop-deps
    echo -e "${GREEN}✓ Cleanup complete${NC}"
    exit 0
}

# Function to wait for PostgreSQL
wait_for_postgres() {
    echo -e "${BLUE}→ Waiting for PostgreSQL to be ready...${NC}"
    local max_attempts=30
    local attempt=0

    cd "$BACKEND_DIR"

    while [ $attempt -lt $max_attempts ]; do
        # Check if the postgres container is healthy
        if docker compose -f deps/docker-compose.yml ps postgres | grep -q "healthy"; then
            echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
            cd "$ROOT_DIR"
            return 0
        fi

        attempt=$((attempt + 1))
        echo -e "${YELLOW}  Waiting... ($attempt/$max_attempts)${NC}"
        sleep 2
    done

    echo -e "${RED}✗ PostgreSQL failed to start within timeout${NC}"
    cd "$ROOT_DIR"
    return 1
}

# Main execution
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   Initializing Development Environment   ║${NC}"
echo -e "${GREEN}╔════════════════════════════════════════╝${NC}"
echo ""

# Phase 0: Initialize Configuration
echo -e "${BLUE}📝 Phase 0: Initializing Configuration${NC}"

# Backend Config
if [ ! -f "$BACKEND_DIR/app.env" ]; then
    echo -e "${YELLOW}  Creating backend config from example...${NC}"
    cp "$BACKEND_DIR/example.env" "$BACKEND_DIR/app.env"
else
    echo -e "${GREEN}  Backend config exists${NC}"
fi

# Frontend Config
FRONTEND_DIR="$ROOT_DIR/next_b2b_starter"
if [ ! -f "$FRONTEND_DIR/.env.local" ]; then
    echo -e "${YELLOW}  Creating frontend config from example...${NC}"
    cp "$FRONTEND_DIR/.env.example" "$FRONTEND_DIR/.env.local"
else
    echo -e "${GREEN}  Frontend config exists${NC}"
fi
echo ""

# Phase 1: Start Docker Dependencies
echo -e "${BLUE}📦 Phase 1: Starting Docker Dependencies${NC}"
cd "$BACKEND_DIR"
make run-deps

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Failed to start Docker dependencies${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker containers started${NC}"
echo ""

# Phase 2: Wait for PostgreSQL
echo -e "${BLUE}🔍 Phase 2: Checking PostgreSQL Health${NC}"
wait_for_postgres
if [ $? -ne 0 ]; then
    cleanup
    exit 1
fi
echo ""

# Phase 3: Run Database Migrations
echo -e "${BLUE}🗄️  Phase 3: Running Database Migrations${NC}"
cd "$BACKEND_DIR"
make migrateup

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Database migrations failed${NC}"
    cleanup
    exit 1
fi
echo -e "${GREEN}✓ Migrations completed${NC}"
echo ""

# Success Output
echo -e "${GREEN}╔══════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   ✅ Environment Ready! 🚀                   ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "You can now run the services in separate terminal tabs:"
echo ""
echo -e "${BLUE}1. Run Backend (with Air hot-reload):${NC}"
echo -e "   cd go-b2b-starter && make dev"
echo ""
echo -e "${BLUE}2. Run Frontend:${NC}"
echo -e "   cd next_b2b_starter && pnpm dev"
echo ""
