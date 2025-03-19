#!/bin/bash

# Exit on error
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}===== HNS Development Setup for WSL (Improved) =====${NC}"

# Command line arguments
SKIP_DOCKER=false
MANUAL_DB=false
GO_PATH=""

# Parse command line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --skip-docker) SKIP_DOCKER=true ;;
        --manual-db) MANUAL_DB=true ;;
        --go-path) GO_PATH="$2"; shift ;;
        --help) 
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --skip-docker    Skip Docker checks and setup"
            echo "  --manual-db      Assume you'll set up the database manually"
            echo "  --go-path PATH   Specify the path to the Go executable"
            echo "  --help           Show this help message"
            exit 0
            ;;
        *) echo "Unknown parameter: $1"; exit 1 ;;
    esac
    shift
done

# Diagnostic information
echo -e "${BLUE}--- System Information ---${NC}"
echo "WSL Distribution: $(cat /etc/os-release | grep -E "^NAME" | cut -d= -f2 | tr -d '"')"
echo "WSL Version: $(if [ -f /proc/version ] && grep -q Microsoft /proc/version; then echo "WSL2"; else echo "WSL1 or native Linux"; fi)"
echo "Shell: $SHELL"
echo "PATH: $PATH"
echo -e "${BLUE}---------------------------${NC}\n"

# Check if Go is installed
GO_EXECUTABLE=""

# Check if Go was specified via command line
if [ -n "$GO_PATH" ]; then
    if [ -x "$GO_PATH" ]; then
        GO_EXECUTABLE="$GO_PATH"
        echo -e "${GREEN}Using specified Go executable at: $GO_EXECUTABLE${NC}"
    else
        echo -e "${RED}Specified Go path '$GO_PATH' is not executable.${NC}"
    fi
else
    # Try standard 'go' command
    if command -v go &> /dev/null; then
        GO_EXECUTABLE="go"
        echo -e "${GREEN}Go found in PATH.${NC}"
    # Try alternative locations
    elif [ -x "/usr/local/go/bin/go" ]; then
        GO_EXECUTABLE="/usr/local/go/bin/go"
        echo -e "${GREEN}Go found at /usr/local/go/bin/go${NC}"
    elif [ -x "$HOME/go/bin/go" ]; then
        GO_EXECUTABLE="$HOME/go/bin/go"
        echo -e "${GREEN}Go found at $HOME/go/bin/go${NC}"
    elif [ -x "$HOME/.go/bin/go" ]; then
        GO_EXECUTABLE="$HOME/.go/bin/go"
        echo -e "${GREEN}Go found at $HOME/.go/bin/go${NC}"
    elif [ -d "/usr/local/go" ]; then
        echo -e "${YELLOW}Go directory found at /usr/local/go but 'go' command not in PATH.${NC}"
        echo -e "${YELLOW}Try adding this to your PATH: export PATH=\$PATH:/usr/local/go/bin${NC}"
    fi
fi

# Verify Go executable and version
if [ -n "$GO_EXECUTABLE" ]; then
    GO_VERSION=$($GO_EXECUTABLE version 2>&1)
    if [[ $GO_VERSION == *"command not found"* ]] || [[ $? -ne 0 ]]; then
        echo -e "${RED}Error running Go executable: $GO_VERSION${NC}"
        GO_EXECUTABLE=""
    else
        echo -e "${GREEN}Go version: $GO_VERSION${NC}"
    fi
fi

# If Go is still not found, provide guidance
if [ -z "$GO_EXECUTABLE" ]; then
    echo -e "${RED}Go is not properly detected in your WSL environment.${NC}"
    echo -e "${YELLOW}Possible solutions:${NC}"
    echo "1. Make sure Go is installed: https://golang.org/doc/install"
    echo "2. Add Go to your PATH: export PATH=\$PATH:/path/to/go/bin"
    echo "3. Restart your shell or run: source ~/.bashrc (or your shell config file)"
    echo "4. Run this script with --go-path to specify the Go executable location"
    echo ""
    echo -e "${YELLOW}To check if Go is installed somewhere, try:${NC}"
    echo "find / -name 'go' -type f -executable 2>/dev/null"
    exit 1
fi

if [ "$SKIP_DOCKER" = false ]; then
    # Check for Docker in different ways
    DOCKER_AVAILABLE=false
    DOCKER_EXECUTABLE=""
    
    # Check if docker command exists
    if command -v docker &> /dev/null; then
        echo -e "${GREEN}Docker command found.${NC}"
        DOCKER_AVAILABLE=true
        DOCKER_EXECUTABLE="docker"
    else
        echo -e "${YELLOW}Docker command not found in PATH.${NC}"
        
        # Check if Docker Desktop is running with WSL integration
        if command -v wsl.exe &> /dev/null && wsl.exe -l --running | grep -q "docker-desktop"; then
            echo -e "${YELLOW}Docker Desktop with WSL integration may be running.${NC}"
            echo -e "${YELLOW}Checking if we can access it...${NC}"
            
            # Try to use the docker context
            if command -v docker.exe &> /dev/null; then
                if docker.exe ps &> /dev/null; then
                    echo -e "${GREEN}Can access Docker through docker.exe${NC}"
                    DOCKER_AVAILABLE=true
                    DOCKER_EXECUTABLE="docker.exe"
                else
                    echo -e "${RED}docker.exe command exists but failed to run.${NC}"
                fi
            fi
        fi
    fi

    # If Docker is still not available, provide guidance
    if [ "$DOCKER_AVAILABLE" = false ]; then
        echo -e "${RED}Docker is not properly configured in your WSL environment.${NC}"
        echo -e "${YELLOW}Options to fix this:${NC}"
        echo "1. Install Docker Desktop with WSL2 integration: https://docs.docker.com/desktop/wsl/"
        echo "2. Install Docker directly in WSL: https://docs.docker.com/engine/install/ubuntu/"
        echo "3. Run this script with --skip-docker to skip Docker setup"
        echo "4. Run this script with --manual-db to assume manual database setup"
        exit 1
    fi

    # Check for docker-compose
    COMPOSE_EXECUTABLE=""
    if command -v docker-compose &> /dev/null; then
        echo -e "${GREEN}docker-compose command found.${NC}"
        COMPOSE_EXECUTABLE="docker-compose"
    elif [ -n "$DOCKER_EXECUTABLE" ] && $DOCKER_EXECUTABLE compose version &> /dev/null; then
        echo -e "${GREEN}Docker Compose plugin is available.${NC}"
        COMPOSE_EXECUTABLE="$DOCKER_EXECUTABLE compose"
    else
        echo -e "${RED}Docker Compose is not installed.${NC}"
        echo -e "${YELLOW}You can run this script with --skip-docker to skip Docker checks.${NC}"
        exit 1
    fi
fi

# Create directories if they don't exist
mkdir -p bin
mkdir -p migrations

if [ "$SKIP_DOCKER" = false ] && [ "$MANUAL_DB" = false ]; then
    # Create docker-compose file if it doesn't exist
    if [ ! -f docker-compose.yml ]; then
        echo -e "${YELLOW}Creating docker-compose.yml file...${NC}"
        cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  postgres:
    image: postgres:16
    container_name: hns-postgres
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: hns
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres-data:
EOF
        echo -e "${GREEN}docker-compose.yml created.${NC}"
    fi

    # Start Docker services
    echo -e "${YELLOW}Starting Docker services...${NC}"
    if [ -n "$COMPOSE_EXECUTABLE" ]; then
        $COMPOSE_EXECUTABLE up -d
    fi
    echo -e "${GREEN}Docker services started.${NC}"

    # Wait for PostgreSQL to be ready
    echo -e "${YELLOW}Waiting for PostgreSQL to be ready...${NC}"
    sleep 5
fi

# Build the migrate tool
echo -e "${YELLOW}Building migration tool...${NC}"
$GO_EXECUTABLE build -o bin/migrate cmd/migrate/main.go
echo -e "${GREEN}Migration tool built.${NC}"

if [ "$MANUAL_DB" = false ]; then
    # Run migrations
    echo -e "${YELLOW}Running database migrations...${NC}"
    ./bin/migrate -up
    echo -e "${GREEN}Database migrations completed.${NC}"
fi

# Build the server
echo -e "${YELLOW}Building HNS server...${NC}"
$GO_EXECUTABLE build -o bin/server cmd/server/main.go
echo -e "${GREEN}HNS server built.${NC}"

echo -e "${GREEN}===== Setup Complete =====${NC}"
echo -e "${YELLOW}To start the server, run:${NC} ./bin/server"
echo -e "${YELLOW}To create a new migration, run:${NC} ./bin/migrate -create \"migration_name\""
echo -e "${YELLOW}To run migrations, run:${NC} ./bin/migrate -up"

if [ "$MANUAL_DB" = true ]; then
    echo -e "${RED}NOTE: You've chosen to set up the database manually.${NC}"
    echo -e "${RED}Make sure to create a PostgreSQL database named 'hns' and update configs/config.yaml accordingly.${NC}"
fi