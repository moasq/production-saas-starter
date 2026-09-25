GO_VERSION := $(shell sed -n 's/^go //p' go.mod)
GO_IMAGE := golang:$(GO_VERSION)-alpine
SQLC_IMAGE := sqlc/sqlc:1.31.1

.PHONY: test vet build run sqlc

test:
	docker run --rm -v "$(CURDIR):/src" -w /src $(GO_IMAGE) go test ./...
vet:
	docker run --rm -v "$(CURDIR):/src" -w /src $(GO_IMAGE) go vet ./...
build:
	docker build -t b2b-starter-backend .
run:
	go run ./cmd/api
sqlc:
	docker run --rm -v "$(CURDIR):/src" -w /src/internal/db/postgres/sqlc $(SQLC_IMAGE) generate
