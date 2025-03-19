#!/bin/bash

# Exit on error
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}===== HNS Database Setup =====${NC}"

# Default values
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="postgres"
DB_NAME="hns"

# Check if PostgreSQL client is installed
if ! command -v psql &> /dev/null; then
    echo -e "${RED}PostgreSQL client (psql) is not installed.${NC}"
    echo -e "${YELLOW}Please install PostgreSQL client:${NC}"
    echo "  For Ubuntu/Debian: sudo apt-get install postgresql-client"
    echo "  For macOS: brew install libpq"
    exit 1
fi

# Create the database
echo -e "${YELLOW}Creating database '${DB_NAME}'...${NC}"

# Create database using PGPASSWORD environment variable
PGPASSWORD="${DB_PASSWORD}" psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d postgres -c "SELECT 1 FROM pg_database WHERE datname = '${DB_NAME}'" | grep -q 1 || \
PGPASSWORD="${DB_PASSWORD}" psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d postgres -c "CREATE DATABASE ${DB_NAME}"

echo -e "${GREEN}Database '${DB_NAME}' created successfully or already exists.${NC}"