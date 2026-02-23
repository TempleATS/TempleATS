package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/temple-ats/TempleATS/internal/db"
)

// TestDB holds a test database connection and its cleanup function.
type TestDB struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
	Cleanup func()
}

// SetupTestDB creates a PostgreSQL container and returns a connected pool.
// It applies the migration schema automatically.
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	ctx := context.Background()

	// Read migration file
	migration, err := os.ReadFile("../../migrations/001_init.sql")
	if err != nil {
		// Try alternate path (tests may run from different directories)
		migration, err = os.ReadFile("../../../migrations/001_init.sql")
		if err != nil {
			t.Fatalf("failed to read migration file: %v", err)
		}
	}

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("temple_ats_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts(), // no init scripts, we'll run migration manually
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Apply migration
	if _, err := pool.Exec(ctx, string(migration)); err != nil {
		t.Fatalf("failed to apply migration: %v", err)
	}

	queries := db.New(pool)

	cleanup := func() {
		pool.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	}

	return &TestDB{
		Pool:    pool,
		Queries: queries,
		Cleanup: cleanup,
	}
}

// CreateTestOrg creates an organization for testing and returns it.
func CreateTestOrg(t *testing.T, q *db.Queries, name, slug string) db.Organization {
	t.Helper()
	ctx := context.Background()
	org, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: name,
		Slug: slug,
	})
	if err != nil {
		t.Fatalf("failed to create test org: %v", err)
	}
	return org
}

// CreateTestUser creates a user for testing and returns it.
func CreateTestUser(t *testing.T, q *db.Queries, email, name, passwordHash, role, orgID string) db.User {
	t.Helper()
	ctx := context.Background()
	user, err := q.CreateUser(ctx, db.CreateUserParams{
		Email:          email,
		Name:           name,
		PasswordHash:   passwordHash,
		Role:           role,
		OrganizationID: orgID,
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}
