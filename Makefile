# Include environment variables
include .envrc
export

# Migration directory
MIGRATION_DIR = ./cmd/migrate/migrations

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make migration name=<name>  - Create a new migration"
	@echo "  make migrate-up             - Run all pending migrations"
	@echo "  make migrate-down [n=1]     - Rollback last n migrations"
	@echo "  make migrate-force v=<ver>  - Force migration to specific version"
	@echo "  make migrate-version        - Show current migration version"
	@echo "  make migrate-drop           - Drop all tables (DANGEROUS!)"
	@echo "  make build                  - Build the application"
	@echo "  make run                    - Run the application"
	@echo "  make dev                    - Run with air (hot reload)"

.PHONY: migration
migration:
	@echo "Creating migration..."
	@migrate create -seq -ext sql -dir $(MIGRATION_DIR) $(name)
	@echo "Migration created."

.PHONY: migrate-up
migrate-up:
	@echo "Running migrations..."
	@migrate -path=$(MIGRATION_DIR) -database=$(DATABASE_URL) up
	@echo "Migrations completed."

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back migrations..."
	@migrate -path=$(MIGRATION_DIR) -database=$(DATABASE_URL) down $(if $(n),$(n),1)
	@echo "Rollback completed."

.PHONY: migrate-force
migrate-force:
	@echo "Forcing migration version to $(v)..."
	@migrate -path=$(MIGRATION_DIR) -database=$(DATABASE_URL) force $(v)
	@echo "Migration forced."

.PHONY: migrate-version
migrate-version:
	@echo "Current migration version:"
	@migrate -path=$(MIGRATION_DIR) -database=$(DATABASE_URL) version

.PHONY: migrate-drop
migrate-drop:
	@echo "⚠️  WARNING: This will drop ALL tables!"
	@read -p "Are you sure? Type 'yes' to continue: " confirm && [ "$$confirm" = "yes" ] || (echo "Aborted." && exit 1)
	@echo "Dropping database..."
	@migrate -path=$(MIGRATION_DIR) -database=$(DATABASE_URL) drop -f
	@echo "Database dropped."

.PHONY: build
build:
	@echo "Building application..."
	@go build -o bin/main cmd/api/main.go
	@echo "Build completed."

.PHONY: run
run:
	@echo "Starting application..."
	@go run cmd/api/main.go

.PHONY: dev
dev:
	@echo "Starting with air (hot reload)..."
	@air

.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "Clean completed."

