package testutil

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/temple-ats/TempleATS/internal/db"
)

var testCounter atomic.Int64

// TestDB holds a test database connection and its cleanup function.
type TestDB struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
	Cleanup func()
}

// SetupTestDB connects to the test database and creates a fresh schema.
// Uses TEST_DATABASE_URL env var. If not set, the test is skipped.
// Each test gets an isolated schema to avoid interference between tests.
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Create a unique schema for this test to ensure isolation
	schemaName := fmt.Sprintf("test_%d_%d", os.Getpid(), testCounter.Add(1))

	if _, err := pool.Exec(ctx, fmt.Sprintf("CREATE SCHEMA %s", schemaName)); err != nil {
		pool.Close()
		t.Fatalf("failed to create test schema: %v", err)
	}

	// Set search_path so all tables are created/queried in this schema
	if _, err := pool.Exec(ctx, fmt.Sprintf("SET search_path TO %s", schemaName)); err != nil {
		pool.Close()
		t.Fatalf("failed to set search_path: %v", err)
	}

	// Read and apply migration
	migration, err := findMigration()
	if err != nil {
		pool.Close()
		t.Fatalf("failed to read migration file: %v", err)
	}

	if _, err := pool.Exec(ctx, string(migration)); err != nil {
		pool.Close()
		t.Fatalf("failed to apply migration: %v", err)
	}

	queries := db.New(pool)

	cleanup := func() {
		pool.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
		pool.Close()
	}

	return &TestDB{
		Pool:    pool,
		Queries: queries,
		Cleanup: cleanup,
	}
}

func findMigration() ([]byte, error) {
	paths := []string{
		"../../migrations/001_init.sql",
		"../../../migrations/001_init.sql",
		"migrations/001_init.sql",
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("migration file not found in any expected path")
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
