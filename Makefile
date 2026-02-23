.PHONY: dev build test migrate sqlc clean

# Development: run Go server with hot reload (requires air) or plain go run
dev:
	DATABASE_URL=$(DATABASE_URL) go run ./cmd/server

# Build: compile React app then embed in Go binary
build: build-web
	go build -o bin/temple-ats ./cmd/server

build-web:
	cd web && npm run build
	rm -rf cmd/server/static
	cp -r web/dist cmd/server/static

# Run all Go tests
test:
	go test ./... -v -count=1

# Run tests with short flag (skip integration tests requiring Docker)
test-short:
	go test ./... -v -short -count=1

# Apply migrations to local database
migrate:
	psql $(DATABASE_URL) -f migrations/001_init.sql

# Generate sqlc code from queries
sqlc:
	sqlc generate

# Clean build artifacts
clean:
	rm -rf bin/ cmd/server/static/
