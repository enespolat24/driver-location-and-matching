.PHONY: test swagger up build down

test: ## Run tests for both services
	@echo "🧪 Running tests..."
	@cd the-driver-location-service && go test ./...
	@cd the-matching-service && go test ./...
	@echo "✅ All tests passed!"

swagger: ## Update swagger docs for both services
	@echo "📚 Updating swagger docs..."
	@cd the-driver-location-service && swag init -g cmd/server/main.go -o docs/
	@cd the-matching-service && swag init -g cmd/server/main.go -o docs/
	@echo "✅ Swagger docs updated!"

up: ## Setup .env files and start docker services
	@echo "🔧 Setting up environment..."
	@cp -n .env.example .env 2>/dev/null || true
	@echo "🐳 Starting services..."
	@docker compose up -d
	@echo "✅ Services started!"

build: ## Setup .env files, build and start docker services
	@echo "🔧 Setting up environment..."
	@cp -n .env.example .env 2>/dev/null || true
	@echo "🔨 Building and starting services..."
	@docker compose up -d --build
	@echo "✅ Services built and started!"

down: ## Stop docker services
	@echo "🛑 Stopping services..."
	@docker compose down
	@echo "✅ Services stopped!"
