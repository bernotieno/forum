#!/bin/bash

# Configuration
IMAGE_NAME="forum-app"
CONTAINER_NAME="forum-container"
PORT=8080

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored messages
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    print_message $RED "Error: Docker is not running"
    exit 1
fi

# Check if container is already running
if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    print_message $YELLOW "Container ${CONTAINER_NAME} is already running"
    print_message $GREEN "You can access the application at http://localhost:${PORT}"
    exit 0
fi

# Check if container exists but is stopped
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    print_message $YELLOW "Starting existing container ${CONTAINER_NAME}..."
    docker start ${CONTAINER_NAME}
    print_message $GREEN "Container started successfully!"
    print_message $GREEN "You can access the application at http://localhost:${PORT}"
    exit 0
fi

# Check if image exists
if ! docker images --format '{{.Repository}}' | grep -q "^${IMAGE_NAME}$"; then
    print_message $YELLOW "Building Docker image ${IMAGE_NAME}..."
    if ! docker build -t ${IMAGE_NAME} .; then
        print_message $RED "Error: Failed to build Docker image"
        exit 1
    fi
    print_message $GREEN "Docker image built successfully!"
fi

# Run the container
print_message $YELLOW "Starting container ${CONTAINER_NAME}..."
if ! docker run -d \
    --name ${CONTAINER_NAME} \
    -p ${PORT}:${PORT} \
    -v forum_data:/app/BackEnd/database/storage \
    ${IMAGE_NAME}; then
    print_message $RED "Error: Failed to start container"
    exit 1
fi

print_message $GREEN "Container started successfully!"
print_message $GREEN "You can access the application at http://localhost:${PORT}" 