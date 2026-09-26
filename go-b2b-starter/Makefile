GO_VERSION := $(shell sed -n 's/^go //p' go.mod)
GO_IMAGE := golang:$(GO_VERSION)-alpine
SQLC_IMAGE := sqlc/sqlc:1.31.1

.PHONY: test test-race test-db vet fmt-check build run sqlc sqlc-check

test:
	docker run --rm -v "$(CURDIR):/src" -w /src $(GO_IMAGE) go test ./...
test-race:
	go test -race ./...
test-db:
	@test -n "$$TEST_DATABASE_URL" || { echo 'Set TEST_DATABASE_URL to a disposable PostgreSQL instance; database tests must not silently skip.' >&2; exit 1; }
	go test -race -count=1 ./internal/db/postgres/...
vet:
	docker run --rm -v "$(CURDIR):/src" -w /src $(GO_IMAGE) go vet ./...
fmt-check:
	docker run --rm -v "$(CURDIR):/src:ro" -w /src $(GO_IMAGE) sh scripts/check-format.sh
build:
	docker build -t b2b-starter-backend .
run:
	go run ./cmd/api
sqlc:
	docker run --rm --user "$$(id -u):$$(id -g)" -v "$(CURDIR):/src" -w /src/internal/db/postgres/sqlc $(SQLC_IMAGE) generate
sqlc-check:
	./scripts/check-sqlc.sh "$(SQLC_IMAGE)"
