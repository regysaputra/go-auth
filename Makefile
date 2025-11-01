# Makefile

include .env
export

# The .PHONY directive tells make that these are not actual files to be built.
.PHONY: run build test clean swag migrate-create migrate-up migrate-down migrate-fix migrate-reset docker-redis

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
	@migrate create -ext sql -dir db/migrations -seq $(args)

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
	@migrate -path db/migrations -database "$(DATABASE_URL)" up
	@echo "Applying all migration..."
	@migrate -path db/migrations -database "$(DATABASE_URL)" down -all
	@echo "Database reset complete"

docker-redis:
	@echo "Running redis..."
	@docker run --name auth-redis -p 6379:6379 -d redis