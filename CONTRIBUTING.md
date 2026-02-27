# Contributing to TempleATS

Thanks for your interest in contributing! This guide will help you get set up.

## Development Setup

### Prerequisites

- Go 1.22+
- Node.js 20+
- PostgreSQL 15+
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html)

### Getting started

```bash
git clone https://github.com/temple-ats/TempleATS.git
cd TempleATS
cp .env.example .env
# Edit .env with your local database credentials

# Create the database and apply migrations
createdb temple_ats
for f in migrations/*.sql; do psql "$DATABASE_URL" -f "$f"; done

# Install frontend dependencies
cd web && npm install && cd ..

# Start development servers
make dev            # Terminal 1 — Go backend on :8080
cd web && npm run dev  # Terminal 2 — React dev server on :5173
```

## Making Changes

### Backend (Go)

- Source code is in `cmd/` and `internal/`
- HTTP handlers go in `internal/handler/`
- Database queries are defined in `queries/*.sql` and generated with `make sqlc` — do not edit files in `internal/db/` directly (except `internal/db/models.go` custom types if needed)
- Run `go vet ./...` before submitting

### Frontend (React + TypeScript)

- Source code is in `web/src/`
- Run `cd web && npm run lint` to check for issues
- Run `cd web && npm run build` to verify the production build compiles

### Database changes

1. Create a new migration file in `migrations/` following the naming pattern: `NNN_description.sql`
2. Write the SQL schema changes
3. Add or update queries in `queries/`
4. Run `make sqlc` to regenerate the Go database layer
5. Commit both the migration and the generated code

## Running Tests

```bash
make test          # All tests (requires PostgreSQL)
make test-short    # Unit tests only (no database needed)
```

## Submitting a Pull Request

1. Fork the repository and create a branch from `main`
2. Make your changes
3. Add tests for new functionality
4. Make sure `make test-short` passes and `make build` succeeds
5. Open a pull request with a clear description of the change

## Code Style

- Go: follow standard `gofmt` formatting
- TypeScript/React: follow the existing ESLint configuration (`cd web && npm run lint`)
- Keep changes focused — one feature or fix per PR

## Reporting Issues

Open an issue on GitHub with:

- A clear description of the problem or feature request
- Steps to reproduce (for bugs)
- Your environment (OS, Go version, Node version, PostgreSQL version)
