# Makefile

include .env
export

# The .PHONY directive tells make that these are not actual files to be built.
.PHONY: run build test clean swag migrate-create migrate-up migrate-down migrate-fix migrate-reset docker-up docker-container-rm docker-volume-rm docker-dev docker-destroy psql ssh dev dev-up dev-down dev-logs dev-clean dev prod prod-up

# Runs the main application
run:
	@echo "Starting the server..."
	@go run ./

# Builds the application binary
build:
	@echo "Building the binary..."
	@go build -o my-api ./cmd/api

# Runs all tests
test:
	@echo "Running tests..."
	@go test ./...

# Cleans up the build artifact
clean:
	@echo "Cleaning up..."
	@rm -f my-api

# Generates Swagger documentation
swag:
	@echo "Generating Swagger docs..."
	@swag init

migrate-create:
	@echo "Create migration file..."
	@migrate create -ext sql -dir internal/infrastructure/db/migrations -seq $(args)

migrate-up:
	@echo "Running database migration up..."
	@migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	@echo "Running database migration down..."
	@migrate -path db/migrations -database "$(DATABASE_URL)" down 1

migrate-fix:
	@echo "Reset a dirty migration"
	@migrate -path db/migrations -database "$(DATABASE_URL)" force $(args)

migrate-reset:
	@echo "Resetting database..."
	@migrate -path internal/infrastructure/db/migrations -database "$(DATABASE_URL)" up
	@echo "Applying all migration..."
	@migrate -path internal/infrastructure/db/migrations -database "$(DATABASE_URL)" down -all
	@echo "Database reset complete"

docker-up:
	@echo "Builds containers..."
	@docker compose up -d

docker-container-rm:
	@echo "Remove all container..."
	@docker rm -v -f $$(docker ps -qa)

docker-volume-rm:
	@echo "Remove all volume..."
	@docker volume prune -a -f

docker-dev:
	@echo "Running docker in dev mode..."
	@docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

docker-destroy:
	@echo "Deep clean docker..."
	@docker system prune -a --force
	@docker volume prune -a --force

psql:
	@echo "Connect to postgresql..."
	@psql postgresql://regy:123@localhost:5433/auth

ssh:
	@ssh -i oci-private.key ubuntu@161.118.206.173

prod:
	@docker compose -f docker-compose.yml build --no-cache

prod-up:
	@docker compose -f docker-compose.yml up

prod-rebuild:
	@docker compose

# Development
dev: ## Start development environment
	@docker compose -f docker-compose.dev.yml build --no-cache

dev-up:
	@docker compose -f docker-compose.dev.yml up

dev-down: ## Stop development environment
	@docker compose -f docker-compose.dev.yml down

dev-logs: ## View development logs
	@docker compose -f docker-compose.dev.yml logs -f

dev-clean: ## Clean development data and rebuild
	@docker compose -f docker-compose.dev.yml down -v
	@docker compose -f docker-compose.dev.yml up --build

