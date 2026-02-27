.PHONY: dev build test migrate seed sqlc clean

# Development: run Go server with hot reload (requires air) or plain go run
dev:
	DATABASE_URL=$(DATABASE_URL) JWT_SECRET=$(JWT_SECRET) go run ./cmd/server

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

# Apply all migrations in order
migrate:
	@for f in migrations/*.sql; do echo "Applying $$f..."; psql $(DATABASE_URL) -f "$$f"; done

# Seed the database with a demo organization and admin user
seed:
	DATABASE_URL=$(DATABASE_URL) go run ./cmd/seed

# Generate sqlc code from queries
sqlc:
	sqlc generate

# Clean build artifacts
clean:
	rm -rf bin/ cmd/server/static/
