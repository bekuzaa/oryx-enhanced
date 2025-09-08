#!/bin/bash

# Enhanced Oryx Docker Build and Run Script
# This script builds and runs the Enhanced Oryx container with all new features

set -e

echo "🚀 Enhanced Oryx Docker Build and Run Script"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    print_success "Docker is running"
}

# Create necessary directories
create_directories() {
    print_status "Creating necessary directories..."
    
    mkdir -p data logs config hls monitoring/prometheus monitoring/grafana/dashboards monitoring/grafana/datasources
    
    print_success "Directories created"
}

# Build Docker image
build_image() {
    print_status "Building Enhanced Oryx Docker image..."
    
    if docker build -f Dockerfile.enhanced -t enhanced-oryx:latest .; then
        print_success "Docker image built successfully"
    else
        print_error "Failed to build Docker image"
        exit 1
    fi
}

# Run with Docker Compose
run_with_compose() {
    print_status "Starting Enhanced Oryx with Docker Compose..."
    
    if docker-compose -f docker-compose.enhanced.yml up -d; then
        print_success "Enhanced Oryx started successfully"
    else
        print_error "Failed to start Enhanced Oryx"
        exit 1
    fi
}

# Run standalone container
run_standalone() {
    print_status "Starting Enhanced Oryx standalone container..."
    
    # Stop existing container if running
    docker stop enhanced-oryx 2>/dev/null || true
    docker rm enhanced-oryx 2>/dev/null || true
    
    # Run container
    docker run -d \
        --name enhanced-oryx \
        --restart unless-stopped \
        -p 2022:2022 \
        -p 1935:1935 \
        -p 8080:8080 \
        -p 1985:1985 \
        -p 10080:10080 \
        -p 10081:10081 \
        -p 10082:10082 \
        -p 80:80 \
        -v "$(pwd)/data:/app/data" \
        -v "$(pwd)/logs:/app/logs" \
        -v "$(pwd)/config:/app/config" \
        -v "$(pwd)/hls:/app/objs/nginx/html/hls" \
        -e REDIS_ADDR=localhost:6379 \
        -e SRS_CONFIG=/app/config/srs.conf \
        -e ORYX_LOG_LEVEL=info \
        -e ORYX_ENABLE_HLS_INPUT=true \
        -e ORYX_ENABLE_SRT_INPUT=true \
        -e ORYX_ENABLE_BYPASS_TRANSCODE=true \
        -e ORYX_ENABLE_MONITORING=true \
        enhanced-oryx:latest
    
    print_success "Enhanced Oryx container started"
}

# Start Redis separately
start_redis() {
    print_status "Starting Redis container..."
    
    docker run -d \
        --name oryx-redis \
        --restart unless-stopped \
        -p 6379:6379 \
        -v redis-data:/data \
        redis:7-alpine \
        redis-server --appendonly yes --bind 0.0.0.0
    
    print_success "Redis started"
}

# Show status
show_status() {
    print_status "Container status:"
    docker ps --filter "name=enhanced-oryx" --filter "name=oryx-redis"
    
    echo ""
    print_status "Service URLs:"
    echo "  Oryx HTTP API:     http://localhost:2022"
    echo "  RTMP:              rtmp://localhost:1935"
    echo "  HLS/HTTP-FLV:      http://localhost:8080"
    echo "  SRS HTTP API:      http://localhost:1985"
    echo "  SRT (with StreamID):    localhost:10080"
    echo "  SRT (no StreamID, 1):   localhost:10081"
    echo "  SRT (no StreamID, 2):   localhost:10082"
    echo "  Nginx:             http://localhost:80"
    echo "  Redis:             localhost:6379"
    
    echo ""
    print_status "Logs:"
    echo "  Oryx logs:         docker logs enhanced-oryx"
    echo "  Redis logs:        docker logs oryx-redis"
}

# Stop containers
stop_containers() {
    print_status "Stopping containers..."
    
    docker stop enhanced-oryx oryx-redis 2>/dev/null || true
    docker rm enhanced-oryx oryx-redis 2>/dev/null || true
    
    print_success "Containers stopped"
}

# Clean up
cleanup() {
    print_status "Cleaning up..."
    
    docker rmi enhanced-oryx:latest 2>/dev/null || true
    docker volume rm redis-data 2>/dev/null || true
    
    print_success "Cleanup completed"
}

# Main menu
show_menu() {
    echo ""
    echo "Choose an option:"
    echo "1) Build and run with Docker Compose (recommended)"
    echo "2) Build and run standalone container"
    echo "3) Build image only"
    echo "4) Start containers only"
    echo "5) Stop containers"
    echo "6) Show status"
    echo "7) Clean up everything"
    echo "8) Exit"
    echo ""
    read -p "Enter your choice (1-8): " choice
    
    case $choice in
        1)
            check_docker
            create_directories
            build_image
            run_with_compose
            show_status
            ;;
        2)
            check_docker
            create_directories
            build_image
            start_redis
            run_standalone
            show_status
            ;;
        3)
            check_docker
            build_image
            ;;
        4)
            check_docker
            run_with_compose
            show_status
            ;;
        5)
            stop_containers
            ;;
        6)
            show_status
            ;;
        7)
            stop_containers
            cleanup
            ;;
        8)
            print_success "Goodbye!"
            exit 0
            ;;
        *)
            print_error "Invalid choice. Please try again."
            show_menu
            ;;
    esac
}

# Check if script is run with arguments
if [ $# -eq 0 ]; then
    show_menu
else
    case $1 in
        "build")
            check_docker
            build_image
            ;;
        "run")
            check_docker
            run_with_compose
            show_status
            ;;
        "stop")
            stop_containers
            ;;
        "status")
            show_status
            ;;
        "cleanup")
            stop_containers
            cleanup
            ;;
        *)
            print_error "Usage: $0 [build|run|stop|status|cleanup]"
            print_error "Or run without arguments for interactive menu"
            exit 1
            ;;
    esac
fi
