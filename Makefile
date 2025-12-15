.PHONY: help dev-up dev-down dev-logs dev-build dev-restart dev-shell prod-up prod-down prod-logs prod-restart prod-shell db-migrate db-shell redis-shell clean setup-dev setup-prod ssl-dev ssl-prod test lint

# Default target
.DEFAULT_GOAL := help

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)Available commands:$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(YELLOW)Quick start:$(NC)"
	@echo "  Development: make setup-dev && make dev-up"
	@echo "  Production:  make setup-prod && make prod-up"

# ============================================
# DEVELOPMENT COMMANDS
# ============================================

dev-up: ## Start development environment
	@echo "$(BLUE)🚀 Starting development environment...$(NC)"
	docker compose -f docker-compose.dev.yml up
	@echo "$(GREEN)✅ Development environment started!$(NC)"
	@echo "$(YELLOW)📍 Access your app at: https://localhost$(NC)"

dev-down: ## Stop development environment
	@echo "$(BLUE)🛑 Stopping development environment...$(NC)"
	docker compose -f docker-compose.dev.yml down
	@echo "$(GREEN)✅ Development environment stopped!$(NC)"

dev-logs: ## View development logs (follow mode)
	@echo "$(BLUE)📋 Viewing development logs... (Ctrl+C to exit)$(NC)"
	docker compose -f docker-compose.yml -f docker-compose.dev.yml logs -f

dev-logs-app: ## View application logs only
	docker compose -f docker-compose.yml -f docker-compose.dev.yml logs -f app

dev-logs-nginx: ## View nginx logs only
	docker compose -f docker-compose.yml -f docker-compose.dev.yml logs -f nginx

dev-build: ## Rebuild development containers
	@echo "$(BLUE)🔨 Building development containers...$(NC)"
	docker compose -f docker-compose.dev.yml build --no-cache
	@echo "$(GREEN)✅ Build complete!$(NC)"

dev-restart: ## Restart development environment
	@echo "$(BLUE)🔄 Restarting development environment...$(NC)"
	docker compose -f docker-compose.yml -f docker-compose.dev.yml restart
	@echo "$(GREEN)✅ Development environment restarted!$(NC)"

dev-shell: ## Open shell in development app container
	@echo "$(BLUE)🐚 Opening shell in app container...$(NC)"
	docker exec -it app sh

dev-ps: ## Show running development containers
	docker compose -f docker-compose.yml -f docker-compose.dev.yml ps

dev-pull: ## Pull latest images for development
	@echo "$(BLUE)⬇️  Pulling latest images...$(NC)"
	docker compose -f docker-compose.yml -f docker-compose.dev.yml pull

# ============================================
# PRODUCTION COMMANDS
# ============================================

prod-up: ## Start production environment
	@echo "$(BLUE)🚀 Starting production environment...$(NC)"
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
	@echo "$(GREEN)✅ Production environment started!$(NC)"
	@echo "$(YELLOW)📍 Access your app at: https://edvora.space$(NC)"

prod-down: ## Stop production environment
	@echo "$(BLUE)🛑 Stopping production environment...$(NC)"
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml down
	@echo "$(GREEN)✅ Production environment stopped!$(NC)"

prod-logs: ## View production logs (follow mode)
	@echo "$(BLUE)📋 Viewing production logs... (Ctrl+C to exit)$(NC)"
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml logs -f

prod-logs-app: ## View production application logs only
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml logs -f app

prod-logs-nginx: ## View production nginx logs only
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml logs -f nginx

prod-restart: ## Restart production environment
	@echo "$(BLUE)🔄 Restarting production environment...$(NC)"
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml restart
	@echo "$(GREEN)✅ Production environment restarted!$(NC)"

prod-shell: ## Open shell in production app container
	@echo "$(BLUE)🐚 Opening shell in app container...$(NC)"
	docker exec -it app-prod sh

prod-ps: ## Show running production containers
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml ps

prod-pull: ## Pull latest images for production
	@echo "$(BLUE)⬇️  Pulling latest images...$(NC)"
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml pull

# ============================================
# DATABASE COMMANDS
# ============================================

db-migrate: ## Run database migrations
	@echo "$(BLUE)⚙️  Running database migrations...$(NC)"
	docker-compose run --rm migrate
	@echo "$(GREEN)✅ Migrations complete!$(NC)"

db-migrate-down: ## Rollback last migration
	@echo "$(YELLOW)⚠️  Rolling back last migration...$(NC)"
	docker-compose run --rm migrate -path /migrations -database "$$DATABASE_URL" down 1

db-migrate-create: ## Create new migration (usage: make db-migrate-create name=add_users_table)
	@if [ -z "$(name)" ]; then \
		echo "$(RED)❌ Error: Please provide migration name$(NC)"; \
		echo "Usage: make db-migrate-create name=add_users_table"; \
		exit 1; \
	fi
	@echo "$(BLUE)📝 Creating migration: $(name)$(NC)"
	docker run --rm -v $(PWD)/internal/infrastructure/db/migrations:/migrations migrate/migrate \
		create -ext sql -dir /migrations -seq $(name)
	@echo "$(GREEN)✅ Migration files created!$(NC)"

db-shell: ## Open PostgreSQL shell
	@echo "$(BLUE)🐘 Opening PostgreSQL shell...$(NC)"
	@if docker ps | grep -q db-prod; then \
		docker exec -it db-prod psql -U $${POSTGRES_USER:-regy} -d $${POSTGRES_DB:-auth}; \
	else \
		docker exec -it db psql -U $${POSTGRES_USER:-regy} -d $${POSTGRES_DB:-auth}; \
	fi

db-backup: ## Backup database (creates backup.sql)
	@echo "$(BLUE)💾 Creating database backup...$(NC)"
	@if docker ps | grep -q db-prod; then \
		docker exec db-prod pg_dump -U $${POSTGRES_USER:-regy} -d $${POSTGRES_DB:-auth} > backup_$$(date +%Y%m%d_%H%M%S).sql; \
	else \
		docker exec db pg_dump -U $${POSTGRES_USER:-regy} -d $${POSTGRES_DB:-auth} > backup_$$(date +%Y%m%d_%H%M%S).sql; \
	fi
	@echo "$(GREEN)✅ Database backup created!$(NC)"

db-restore: ## Restore database from backup (usage: make db-restore file=backup.sql)
	@if [ -z "$(file)" ]; then \
		echo "$(RED)❌ Error: Please provide backup file$(NC)"; \
		echo "Usage: make db-restore file=backup.sql"; \
		exit 1; \
	fi
	@echo "$(YELLOW)⚠️  Restoring database from $(file)...$(NC)"
	@if docker ps | grep -q db-prod; then \
		docker exec -i db-prod psql -U $${POSTGRES_USER:-regy} -d $${POSTGRES_DB:-auth} < $(file); \
	else \
		docker exec -i db psql -U $${POSTGRES_USER:-regy} -d $${POSTGRES_DB:-auth} < $(file); \
	fi
	@echo "$(GREEN)✅ Database restored!$(NC)"

# ============================================
# REDIS COMMANDS
# ============================================

redis-shell: ## Open Redis CLI
	@echo "$(BLUE)🔴 Opening Redis CLI...$(NC)"
	@if docker ps | grep -q redis-prod; then \
		docker exec -it redis-prod redis-cli; \
	else \
		docker exec -it redis redis-cli; \
	fi

redis-flush: ## Flush all Redis data (⚠️  WARNING: Deletes all data!)
	@echo "$(RED)⚠️  WARNING: This will delete all Redis data!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		if docker ps | grep -q redis-prod; then \
			docker exec redis-prod redis-cli FLUSHALL; \
		else \
			docker exec redis redis-cli FLUSHALL; \
		fi; \
		echo "$(GREEN)✅ Redis flushed!$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled.$(NC)"; \
	fi

redis-monitor: ## Monitor Redis commands in real-time
	@echo "$(BLUE)👀 Monitoring Redis... (Ctrl+C to exit)$(NC)"
	@if docker ps | grep -q redis-prod; then \
		docker exec -it redis-prod redis-cli MONITOR; \
	else \
		docker exec -it redis redis-cli MONITOR; \
	fi

# ============================================
# SSL/HTTPS COMMANDS
# ============================================

ssl-dev: ## Setup local HTTPS certificates for development
	@echo "$(BLUE)🔒 Setting up local HTTPS certificates...$(NC)"
	@chmod +x scripts/setup-local-https.sh
	@./scripts/setup-local-https.sh
	@echo "$(GREEN)✅ Local HTTPS setup complete!$(NC)"

ssl-prod: ## Obtain Let's Encrypt SSL certificate for production
	@echo "$(BLUE)🔒 Obtaining SSL certificate from Let's Encrypt...$(NC)"
	@read -p "Enter your domain (e.g., edvora.space): " domain; \
	read -p "Enter your email: " email; \
	chmod +x scripts/obtain-ssl-cert.sh; \
	./scripts/obtain-ssl-cert.sh $$domain $$email

ssl-renew: ## Manually renew SSL certificate
	@echo "$(BLUE)🔄 Renewing SSL certificate...$(NC)"
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml run --rm certbot renew
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml exec nginx nginx -s reload
	@echo "$(GREEN)✅ Certificate renewed!$(NC)"

ssl-check: ## Check SSL certificate expiration
	@echo "$(BLUE)🔍 Checking SSL certificate...$(NC)"
	@if [ -d "nginx/certs/live" ]; then \
		docker-compose -f docker-compose.yml -f docker-compose.prod.yml run --rm certbot certificates; \
	else \
		echo "$(YELLOW)No Let's Encrypt certificates found.$(NC)"; \
	fi

# ============================================
# TESTING & QUALITY COMMANDS
# ============================================

test: ## Run Go tests
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	go test -v ./...
	@echo "$(GREEN)✅ Tests complete!$(NC)"

test-coverage: ## Run tests with coverage
	@echo "$(BLUE)🧪 Running tests with coverage...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(NC)"

lint: ## Run Go linter
	@echo "$(BLUE)🔍 Running linter...$(NC)"
	golangci-lint run ./...
	@echo "$(GREEN)✅ Linting complete!$(NC)"

fmt: ## Format Go code
	@echo "$(BLUE)✨ Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✅ Code formatted!$(NC)"

# ============================================
# SETUP COMMANDS
# ============================================

setup-dev: ## Complete development environment setup
	@echo "$(BLUE)🔧 Setting up development environment...$(NC)"
	@echo "$(YELLOW)Step 1/4: Installing mkcert and generating certificates...$(NC)"
	@chmod +x scripts/setup-local-https.sh
	@./scripts/setup-local-https.sh
	@echo "$(YELLOW)Step 2/4: Creating .env file...$(NC)"
	@if [ ! -f .env ]; then \
		cp .env.development .env; \
		echo "$(GREEN)✅ .env file created from .env.development$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  .env file already exists, skipping...$(NC)"; \
	fi
	@echo "$(YELLOW)Step 3/4: Creating necessary directories...$(NC)"
	@mkdir -p nginx/certs nginx/certbot-webroot pkg/geoip templates
	@echo "$(YELLOW)Step 4/4: Starting development environment...$(NC)"
	@make dev-up
	@echo ""
	@echo "$(GREEN)✅ Development environment setup complete!$(NC)"
	@echo "$(BLUE)📍 Access your app at: https://localhost$(NC)"

setup-prod: ## Setup production environment (run on server)
	@echo "$(BLUE)🔧 Setting up production environment...$(NC)"
	@echo "$(YELLOW)Step 1/3: Creating directories...$(NC)"
	@mkdir -p nginx/certs nginx/certbot-webroot pkg/geoip templates
	@echo "$(YELLOW)Step 2/3: Creating .env file...$(NC)"
	@if [ ! -f .env ]; then \
		cp .env.production .env; \
		echo "$(GREEN)✅ .env file created from .env.production$(NC)"; \
		echo "$(RED)⚠️  IMPORTANT: Edit .env with your production values!$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  .env file already exists, skipping...$(NC)"; \
	fi
	@echo "$(YELLOW)Step 3/3: Instructions...$(NC)"
	@echo ""
	@echo "$(BLUE)Next steps:$(NC)"
	@echo "  1. Edit .env with your production values: nano .env"
	@echo "  2. Start production: make prod-up"
	@echo "  3. Get SSL certificate: make ssl-prod"

# ============================================
# CLEANUP COMMANDS
# ============================================

clean: ## Remove all containers, volumes, and images
	@echo "$(RED)⚠️  WARNING: This will remove all containers, volumes, and unused images!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "$(BLUE)🧹 Cleaning up...$(NC)"; \
		docker-compose -f docker-compose.yml -f docker-compose.dev.yml down -v; \
		docker-compose -f docker-compose.yml -f docker-compose.prod.yml down -v; \
		docker system prune -f; \
		echo "$(GREEN)✅ Cleanup complete!$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled.$(NC)"; \
	fi

clean-dev: ## Remove development containers and volumes
	@echo "$(BLUE)🧹 Cleaning development environment...$(NC)"
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml down -v
	@echo "$(GREEN)✅ Development environment cleaned!$(NC)"

clean-prod: ## Remove production containers and volumes
	@echo "$(RED)⚠️  WARNING: This will remove production containers and volumes!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose -f docker-compose.yml -f docker-compose.prod.yml down -v; \
		echo "$(GREEN)✅ Production environment cleaned!$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled.$(NC)"; \
	fi

clean-images: ## Remove unused Docker images
	@echo "$(BLUE)🧹 Removing unused images...$(NC)"
	docker image prune -af
	@echo "$(GREEN)✅ Unused images removed!$(NC)"

# ============================================
# UTILITY COMMANDS
# ============================================

status: ## Show status of all services
	@echo "$(BLUE)📊 Development Services:$(NC)"
	@docker-compose -f docker-compose.yml -f docker-compose.dev.yml ps 2>/dev/null || echo "Not running"
	@echo ""
	@echo "$(BLUE)📊 Production Services:$(NC)"
	@docker-compose -f docker-compose.yml -f docker-compose.prod.yml ps 2>/dev/null || echo "Not running"

health: ## Check health of running services
	@echo "$(BLUE)🏥 Checking service health...$(NC)"
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

stats: ## Show Docker container stats
	@echo "$(BLUE)📈 Container statistics... (Ctrl+C to exit)$(NC)"
	docker stats

nginx-test: ## Test nginx configuration
	@echo "$(BLUE)🔍 Testing nginx configuration...$(NC)"
	@if docker ps | grep -q nginx-prod; then \
		docker exec nginx-prod nginx -t; \
	elif docker ps | grep -q nginx-dev; then \
		docker exec nginx-dev nginx -t; \
	else \
		echo "$(RED)❌ Nginx container not running$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ Nginx configuration is valid!$(NC)"

nginx-reload: ## Reload nginx configuration
	@echo "$(BLUE)🔄 Reloading nginx...$(NC)"
	@if docker ps | grep -q nginx-prod; then \
		docker exec nginx-prod nginx -s reload; \
	elif docker ps | grep -q nginx-dev; then \
		docker exec nginx-dev nginx -s reload; \
	else \
		echo "$(RED)❌ Nginx container not running$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ Nginx reloaded!$(NC)"

update: ## Pull latest code and restart services
	@echo "$(BLUE)🔄 Updating application...$(NC)"
	git pull origin main
	@if docker ps | grep -q prod; then \
		make prod-down && make prod-up; \
	else \
		make dev-down && make dev-up; \
	fi
	@echo "$(GREEN)✅ Application updated!$(NC)"