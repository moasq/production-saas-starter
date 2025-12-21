COMPOSE_FILE := deps/docker-compose.yml
MIGRATION_PATH ?= schema/migration
POSTGRES_HOST ?= localhost
POSTGRES_PORT ?= 5432
POSTGRES_DB ?= mydatabase
POSTGRES_USER ?= user
POSTGRES_PASSWORD ?= password
CONTAINER_NAME ?= deps-postgis-1
MIGRATION_NAME ?= init_schema
MIGRATION_DIR ?= ./internal/db/postgres/sqlc/migrations
SQLC_DIR ?= internal/db/postgres/sqlc

# Start the necessary docker containers
run-deps:
	docker compose -f $(COMPOSE_FILE) up --build -d

# Stop and remove docker containers
stop-deps:
	docker compose -f $(COMPOSE_FILE) down -v

# Create a new database migration file
create-migration:
	@docker compose -f $(COMPOSE_FILE) run --rm cli migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(MIGRATION_NAME)
	@echo "Migration created in $(MIGRATION_DIR) with name $(MIGRATION_NAME)"

# Apply all up migrations
# Apply all up migrations
migrateup:
	@docker compose -f $(COMPOSE_FILE) run --rm cli migrate -path $(MIGRATION_DIR) -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" -verbose up

# Apply all down migrations
migratedown:
	@docker compose -f $(COMPOSE_FILE) run --rm cli migrate -path $(MIGRATION_DIR) -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" -verbose down

sqlc:
	@docker compose -f $(COMPOSE_FILE) run --rm -w /workspace/$(SQLC_DIR) cli sqlc generate

# Create a new module
create-module:
	bash scripts/create_module.sh $(type) $(name)
	if [ "$(db)" = "postgres" ]; then bash scripts/setup_db.sh $(type) $(name); fi

# Run the server
server:
	go run ./cmd/api/main.go

# build the app
build:
	go build -o bin/api ./cmd/api/main.go

# install dependencies
deps:
	go mod tidy

# swagger
swagger:
	@docker compose -f $(COMPOSE_FILE) run --rm cli swag init -g cmd/api/main.go -d . --parseDependency --parseInternal -o internal/docs/gen

# Run the server with Air (Live Reload)
dev:
	@docker compose -f $(COMPOSE_FILE) run --rm -T --service-ports cli air 

reload-profile:
	@source ~/.bashrc

test:
	@bash scripts/run_tests_with_coverage.sh

# Clear RBAC and JWKS caches from Redis
clear-rbac-cache:
	@echo "Clearing RBAC and JWKS Redis caches..."
	@redis-cli DEL "stytch:rbac:policy" || echo "  ✗ Failed to clear stytch:rbac:policy (may not exist)"
	@redis-cli DEL "stytch:jwks:cache" || echo "  ✗ Failed to clear stytch:jwks:cache (may not exist)"
	@echo "✓ Cache clearing attempted (caches will auto-expire if not manually cleared)"


.PHONY: \
    build \
    clear-rbac-cache \
    create-migration \
    create-module \
    create-seed-country \
    deps \
    generate-seed-file \
    generate-changed-seed-file \
	generate-migrations-file \
	generate-down-migrations-file \
    migratedown \
    migrateup \
    push-to-do \
    reload-profile \
    run-deps \
    seed-db \
    server \
    sonar-scanner \
    sqlc \
    swagger \
    test
