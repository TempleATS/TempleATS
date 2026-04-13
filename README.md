# TempleATS

> Recruiter sees TempleOS - *"It doesn't look very good"*
>
> Senior SWE - *"Let's lock you in a room with a computer and see how far you get"*

A modern, self-hosted Applicant Tracking System built with Go and React. Manage your hiring pipeline from requisition to offer — organize candidates, schedule interviews, collaborate with your team, and publish a public careers page.

## Features

- **Requisition management** — create and track hiring needs with assigned recruiters and hiring managers
- **Job postings** — publish jobs with structured descriptions and attach them to requisitions
- **Candidate pipeline** — Kanban board to move candidates through customizable stages
- **Public careers page** — branded job listing page where candidates can apply directly
- **Interview scheduling** — schedule interviews with Google Calendar integration
- **Interview feedback** — structured feedback collection from interviewers
- **Email notifications** — configurable per-stage email templates with custom SMTP
- **Employee referrals** — referral tracking with shareable links
- **Hiring packets** — generate PDF summaries for final review
- **Team management** — invite members with role-based access control
- **Multi-tenant** — each organization has isolated data

### Roles

| Role | Capabilities |
|------|-------------|
| Super Admin | Full access including org settings and SMTP configuration |
| Admin | Create/manage requisitions, jobs, and team members |
| Recruiter | Manage candidates, schedule interviews, send emails |
| Hiring Manager | Move candidates through stages, assign interviewers |
| Interviewer | View assigned candidates, submit feedback |

## Tech Stack

- **Backend:** Go, Chi router, pgx (PostgreSQL), sqlc
- **Frontend:** React 19, TypeScript, Tailwind CSS, Vite
- **Database:** PostgreSQL
- **Auth:** JWT (cookie-based)

## Prerequisites

- Go 1.22+
- Node.js 20+
- PostgreSQL 15+
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) (only if modifying queries)

## Quick Start

### 1. Clone and configure

```bash
git clone https://github.com/temple-ats/TempleATS.git
cd TempleATS
cp .env.example .env
# Edit .env with your database credentials and a random JWT_SECRET
```

### 2. Create the database

```bash
createdb temple_ats
```

### 3. Install frontend dependencies

```bash
cd web && npm install && cd ..
```

### 4. Run in development mode

Start the Go backend and React dev server separately:

```bash
# Terminal 1 — backend
make dev

# Terminal 2 — frontend (with hot reload, proxies API to :8080)
cd web && npm run dev
```

The app will be available at `http://localhost:5173`. Database migrations run automatically on startup.

Sign up to create your first organization and admin account, or seed demo data:

```bash
make seed
```

This creates a "Demo Company" org with login credentials `admin@example.com` / `password`.

### Production Build

Build the React frontend and embed it into the Go binary:

```bash
make build
```

This produces `bin/temple-ats` — a single binary that serves both the API and the frontend.

Run it:

```bash
DATABASE_URL=postgresql://... JWT_SECRET=your-secret ./bin/temple-ats
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `JWT_SECRET` | Yes | — | Secret key for signing auth tokens |
| `PORT` | No | `8080` | HTTP server port |
| `UPLOAD_DIR` | No | `./uploads` | Directory for resume file uploads |

## Project Structure

```
cmd/
  server/          # Application entry point, routing, auto-migration
  seed/            # Seed demo data (org + admin user)
internal/
  auth/            # JWT token generation and validation
  db/              # sqlc-generated database layer
  email/           # SMTP email service
  handler/         # HTTP request handlers
  migrate/         # Auto-migration runner
  middleware/      # Auth and RBAC middleware
  calendar/        # Google Calendar integration
migrations/        # PostgreSQL schema migrations (applied in order)
queries/           # SQL queries (sqlc source files)
web/               # React + TypeScript frontend
```

## Development

### Makefile targets

| Command | Description |
|---------|-------------|
| `make dev` | Run the Go server |
| `make build` | Build frontend + Go binary |
| `make build-web` | Build React app only |
| `make test` | Run all tests |
| `make test-short` | Run tests (skip integration) |
| `make migrate` | Apply all migrations via psql |
| `make seed` | Create demo org + admin user |
| `make sqlc` | Regenerate Go code from SQL queries |
| `make clean` | Remove build artifacts |

### Modifying the database layer

TempleATS uses [sqlc](https://sqlc.dev) to generate type-safe Go code from SQL:

1. Edit or add query files in `queries/`
2. Run `make sqlc`
3. The generated code appears in `internal/db/`

### Running tests

```bash
make test          # all tests (requires a running PostgreSQL)
make test-short    # unit tests only
```

## Public Careers Page

Each organization gets a public careers page at `/careers/{org-slug}` that lists all open jobs. Candidates can view job details and apply directly — no account needed.

To integrate this into your company's existing website:

**Direct link** — add a "View Open Positions" link pointing to your instance:
```
https://your-ats-domain.com/careers/your-org-slug
```

**Iframe embed** — embed the careers page inline:
```html
<iframe src="https://your-ats-domain.com/careers/your-org-slug"
        width="100%" height="800" frameborder="0"></iframe>
```

**API integration** — build a fully custom UI using the public endpoints:
```
GET  /api/careers/{orgSlug}                    # List open jobs
GET  /api/careers/{orgSlug}/jobs/{jobId}       # Job details
POST /api/careers/{orgSlug}/jobs/{jobId}/apply  # Submit application
```

## Recent Improvements

- **Full-text search** — jobs and requisitions searchable via PostgreSQL `tsvector` + `pg_trgm` fuzzy matching
- **S3 file storage** — resumes can be stored on S3-compatible storage (AWS, MinIO, DO Spaces) via `STORAGE_TYPE=s3`
- **Rate limiting** — per-IP token bucket rate limiting with configurable RPS/burst; stricter limits on auth routes
- **Audit logging** — all mutations (create/update/delete) logged to `audit_logs` table with user, entity, and details
- **Configurable CORS** — set `CORS_ORIGINS` env var for production domains (comma-separated)
- **SSO via OIDC** — authenticate with any OIDC provider (Google, Okta, etc.) via `SSO_ISSUER` env vars
- **Frontend tests** — vitest + testing-library test suite for Jobs and Requisitions pages

## License

[MIT](LICENSE)
