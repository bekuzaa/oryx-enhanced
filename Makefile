# Enhanced Oryx Makefile
# Provides convenient commands for building and running Enhanced Oryx

.PHONY: help build run stop status logs clean docker-build docker-run docker-stop docker-logs docker-clean

# Default target
help:
	@echo "Enhanced Oryx - Available Commands:"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker-build    Build Enhanced Oryx Docker image"
	@echo "  docker-run      Run Enhanced Oryx with Docker Compose"
	@echo "  docker-stop     Stop all containers"
	@echo "  docker-logs     Show container logs"
	@echo "  docker-clean    Clean up containers and images"
	@echo ""
	@echo "Development Commands:"
	@echo "  build           Build Go binary"
	@echo "  run             Run Go binary directly"
	@echo "  stop            Stop running processes"
	@echo "  status          Show running status"
	@echo "  logs            Show application logs"
	@echo "  clean           Clean build artifacts"
	@echo ""
	@echo "Utility Commands:"
	@echo "  setup           Setup development environment"
	@echo "  test            Run tests"
	@echo "  fmt             Format Go code"
	@echo "  lint            Lint Go code"
	@echo "  deps            Download Go dependencies"

# Docker Commands
docker-build:
	@echo "🐳 Building Enhanced Oryx Docker image..."
	docker build -f Dockerfile.enhanced -t enhanced-oryx:latest .
	@echo "✅ Docker image built successfully"

docker-run:
	@echo "🚀 Starting Enhanced Oryx with Docker Compose..."
	docker-compose -f docker-compose.enhanced.yml up -d
	@echo "✅ Enhanced Oryx started successfully"
	@echo "📊 Check status with: make status"

docker-stop:
	@echo "🛑 Stopping Enhanced Oryx containers..."
	docker-compose -f docker-compose.enhanced.yml down
	@echo "✅ Containers stopped"

docker-logs:
	@echo "📋 Container logs:"
	docker-compose -f docker-compose.enhanced.yml logs -f

docker-clean:
	@echo "🧹 Cleaning up Docker resources..."
	docker-compose -f docker-compose.enhanced.yml down -v
	docker rmi enhanced-oryx:latest 2>/dev/null || true
	docker system prune -f
	@echo "✅ Cleanup completed"

# Development Commands
build:
	@echo "🔨 Building Enhanced Oryx binary..."
	cd platform && go build -o oryx .
	@echo "✅ Binary built successfully"

run:
	@echo "🏃 Running Enhanced Oryx..."
	cd platform && ./oryx

stop:
	@echo "🛑 Stopping Enhanced Oryx..."
	pkill -f "oryx" || true

status:
	@echo "📊 Enhanced Oryx Status:"
	@echo ""
	@echo "Docker Containers:"
	docker ps --filter "name=enhanced-oryx" --filter "name=oryx-redis" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
	@echo ""
	@echo "Service URLs:"
	@echo "  Oryx HTTP API:     http://localhost:2022"
	@echo "  RTMP:              rtmp://localhost:1935"
	@echo "  HLS/HTTP-FLV:      http://localhost:8080"
	@echo "  SRS HTTP API:      http://localhost:1985"
	@echo "  SRT (with StreamID):    localhost:10080"
	@echo "  SRT (no StreamID, 1):   localhost:10081"
	@echo "  SRT (no StreamID, 2):   localhost:10082"
	@echo "  Nginx:             http://localhost:80"
	@echo "  Redis:             localhost:6379"

logs:
	@echo "📋 Application logs:"
	@if [ -f "logs/oryx.log" ]; then tail -f logs/oryx.log; else echo "No log file found"; fi

clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f platform/oryx platform/oryx.exe
	rm -rf logs/* data/* hls/*
	@echo "✅ Cleanup completed"

# Utility Commands
setup:
	@echo "⚙️ Setting up development environment..."
	mkdir -p data logs config hls monitoring/prometheus monitoring/grafana/dashboards monitoring/grafana/datasources
	@echo "✅ Environment setup completed"

test:
	@echo "🧪 Running tests..."
	cd platform && go test ./...

fmt:
	@echo "🎨 Formatting Go code..."
	cd platform && go fmt ./...

lint:
	@echo "🔍 Linting Go code..."
	cd platform && golangci-lint run

deps:
	@echo "📦 Downloading Go dependencies..."
	cd platform && go mod download

# Quick start commands
quick-start: docker-build docker-run
	@echo "🚀 Quick start completed!"
	@echo "📊 Check status with: make status"
	@echo "📋 View logs with: make docker-logs"

quick-stop: docker-stop
	@echo "🛑 Quick stop completed!"

# Development workflow
dev-setup: setup deps
	@echo "✅ Development environment ready"

dev-build: fmt build
	@echo "✅ Development build completed"

dev-run: dev-build run
	@echo "✅ Development run completed"

# Production commands
prod-build: docker-build
	@echo "🏭 Production build completed"

prod-deploy: prod-build docker-run
	@echo "🚀 Production deployment completed"

prod-stop: docker-stop
	@echo "🛑 Production stopped"

# Monitoring commands
monitoring-start:
	@echo "📊 Starting monitoring stack..."
	docker-compose -f docker-compose.enhanced.yml up -d prometheus grafana
	@echo "✅ Monitoring started"
	@echo "📈 Grafana: http://localhost:3000 (admin/admin)"
	@echo "📊 Prometheus: http://localhost:9090"

monitoring-stop:
	@echo "🛑 Stopping monitoring stack..."
	docker-compose -f docker-compose.enhanced.yml stop prometheus grafana
	@echo "✅ Monitoring stopped"

# Backup and restore
backup:
	@echo "💾 Creating backup..."
	mkdir -p backups
	tar -czf backups/oryx-backup-$(shell date +%Y%m%d-%H%M%S).tar.gz data logs config hls
	@echo "✅ Backup created"

restore:
	@echo "📥 Restoring from backup..."
	@if [ -z "$(BACKUP_FILE)" ]; then echo "Usage: make restore BACKUP_FILE=backup.tar.gz"; exit 1; fi
	tar -xzf $(BACKUP_FILE)
	@echo "✅ Restore completed"

# Help for specific commands
docker-help:
	@echo "🐳 Docker Commands Help:"
	@echo "  make docker-build    - Build Docker image"
	@echo "  make docker-run      - Start containers"
	@echo "  make docker-stop     - Stop containers"
	@echo "  make docker-logs     - View logs"
	@echo "  make docker-clean    - Clean up everything"

dev-help:
	@echo "🔨 Development Commands Help:"
	@echo "  make dev-setup       - Setup environment"
	@echo "  make dev-build       - Build binary"
	@echo "  make dev-run         - Run binary"
	@echo "  make test            - Run tests"
	@echo "  make fmt             - Format code"
	@echo "  make lint            - Lint code"

# Default target
.DEFAULT_GOAL := help
