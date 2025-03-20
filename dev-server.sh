#!/bin/bash

# Exit on error
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}===== HNS Development Server =====${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed. Please install Go first.${NC}"
    exit 1
fi

# Start PostgreSQL if using Docker
if command -v docker &> /dev/null && docker ps &> /dev/null; then
    # Check if Postgres container is running
    if ! docker ps | grep -q "hns-postgres"; then
        echo -e "${YELLOW}Starting PostgreSQL container...${NC}"
        docker-compose up -d postgres
        
        # Wait for PostgreSQL to be ready
        echo -e "${YELLOW}Waiting for PostgreSQL to be ready...${NC}"
        sleep 5
    else
        echo -e "${GREEN}PostgreSQL container is already running.${NC}"
    fi
else
    echo -e "${YELLOW}Docker not available. Make sure PostgreSQL is running manually.${NC}"
fi

# Run migrations
echo -e "${YELLOW}Running database migrations...${NC}"
go run cmd/migrate/main.go -up

# Build and run the server with live reload
if command -v air &> /dev/null; then
    echo -e "${GREEN}Starting server with hot reload using Air...${NC}"
    air
else
    echo -e "${YELLOW}Air not installed. Running server without hot reload.${NC}"
    echo -e "${YELLOW}To install Air for hot reload: go install github.com/cosmtrek/air@latest${NC}"
    go run cmd/server/main.go
fi