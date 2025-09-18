# Makefile for Linier Channel Local Development
# Uses Dockerfile directly (Production-like)

.PHONY: help build run stop logs clean status

# Configuration
IMAGE_NAME = linier-channel
CONTAINER_NAME = linier-channel-app
PORT = 8080

# Local storage paths
UPLOAD_PATH = /Users/herihandoko/Documents/Transcode/input
TRANCODED_PATH = /Users/herihandoko/Documents/Transcode/output
ARCHIVE_PATH = /Users/herihandoko/Documents/Transcode/archive
LOG_PATH = /Users/herihandoko/Documents/Transcode/log

help: ## Show this help message
	@echo "🐳 Linier Channel Local Development"
	@echo "=================================="
	@echo ""
	@echo "Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make build    # Build Docker image"
	@echo "  make run      # Run container"
	@echo "  make logs     # View logs"
	@echo "  make stop     # Stop container"

build: ## Build Docker image
	@echo "🔨 Building Docker image..."
	docker build -t $(IMAGE_NAME) .
	@echo "✅ Image built successfully: $(IMAGE_NAME)"

run: ## Run container (builds if needed)
	@echo "🚀 Running container..."
	@mkdir -p $(UPLOAD_PATH) $(TRANCODED_PATH) $(ARCHIVE_PATH) $(LOG_PATH)
	@docker stop $(CONTAINER_NAME) 2>/dev/null || true
	@docker rm $(CONTAINER_NAME) 2>/dev/null || true
	docker run -d \
		--name $(CONTAINER_NAME) \
		-p $(PORT):8080 \
		-e SERVER_HOST=0.0.0.0 \
		-e SERVER_PORT=8080 \
		-e DB_HOST=host.docker.internal \
		-e DB_PORT=3306 \
		-e DB_USER=root \
		-e DB_PASSWORD=Nd45mulh0! \
		-e DB_NAME=linier_channel \
		-e REDIS_HOST=host.docker.internal \
		-e REDIS_PORT=6379 \
		-e UPLOAD_PATH=/uploads \
		-e TRANCODED_PATH=/transcoded-videos \
		-e ARCHIVE_PATH=/archive \
		-e LOG_PATH=/var/log/linier-channel/app.log \
		-v $(UPLOAD_PATH):/uploads \
		-v $(TRANCODED_PATH):/transcoded-videos \
		-v $(ARCHIVE_PATH):/archive \
		-v $(LOG_PATH):/var/log/linier-channel \
		--add-host=host.docker.internal:host-gateway \
		$(IMAGE_NAME)
	@echo "✅ Container started!"
	@echo "🌐 Application: http://localhost:$(PORT)"
	@echo "📁 Logs: $(LOG_PATH)"

stop: ## Stop container
	@echo "🛑 Stopping container..."
	@docker stop $(CONTAINER_NAME) 2>/dev/null || true
	@echo "✅ Container stopped"

logs: ## View container logs
	@echo "📋 Container logs:"
	docker logs -f $(CONTAINER_NAME)

status: ## Show container status
	@echo "📊 Container status:"
	@docker ps -f name=$(CONTAINER_NAME)
	@echo ""
	@echo "📋 Recent logs:"
	@docker logs --tail 20 $(CONTAINER_NAME)

clean: ## Stop and remove container
	@echo "🧹 Cleaning up..."
	@docker stop $(CONTAINER_NAME) 2>/dev/null || true
	@docker rm $(CONTAINER_NAME) 2>/dev/null || true
	@echo "✅ Cleanup completed"

restart: stop run ## Restart container

# Quick commands
dev: build run ## Build and run (development)
prod: build run ## Build and run (production-like)

# Default target
.DEFAULT_GOAL := help
