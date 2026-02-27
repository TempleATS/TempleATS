package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/temple-ats/TempleATS/internal/db"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	q := db.New(pool)

	// Check if any organizations exist
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM organizations").Scan(&count)
	if err != nil {
		log.Fatalf("failed to check organizations: %v", err)
	}
	if count > 0 {
		fmt.Println("Database already has organizations. Skipping seed.")
		return
	}

	// Create demo organization
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	qtx := q.WithTx(tx)

	org, err := qtx.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: "Demo Company",
		Slug: "demo",
	})
	if err != nil {
		log.Fatalf("failed to create organization: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	user, err := qtx.CreateUser(ctx, db.CreateUserParams{
		Email:          "admin@example.com",
		Name:           "Admin User",
		PasswordHash:   string(hash),
		Role:           "super_admin",
		OrganizationID: org.ID,
	})
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("failed to commit: %v", err)
	}

	fmt.Println("Seed complete!")
	fmt.Printf("  Organization: %s (slug: %s)\n", org.Name, org.Slug)
	fmt.Printf("  Admin user:   %s (%s)\n", user.Email, user.Name)
	fmt.Println()
	fmt.Println("Login with:")
	fmt.Println("  Email:    admin@example.com")
	fmt.Println("  Password: password")
}
