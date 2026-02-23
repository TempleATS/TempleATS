package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/temple-ats/TempleATS/internal/handler"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

//go:embed static
var staticFiles embed.FS

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	srv := handler.NewServer(pool)

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/api/health", srv.Health)

	// Auth routes (public)
	r.Post("/api/auth/signup", srv.Signup)
	r.Post("/api/auth/login", srv.Login)
	r.Post("/api/auth/logout", srv.Logout)

	// Public careers routes (no auth)
	r.Get("/api/careers/{orgSlug}", srv.CareersListJobs)
	r.Get("/api/careers/{orgSlug}/jobs/{jobId}", srv.CareersGetJob)
	r.Post("/api/careers/{orgSlug}/jobs/{jobId}/apply", srv.CareersApply)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(mw.RequireAuth)
		r.Get("/api/auth/me", srv.Me)

		// Requisitions
		r.Get("/api/reqs", srv.ListRequisitions)
		r.Post("/api/reqs", srv.CreateRequisition)
		r.Get("/api/reqs/{reqId}", srv.GetRequisition)
		r.Put("/api/reqs/{reqId}", srv.UpdateRequisition)
		r.Post("/api/reqs/{reqId}/jobs", srv.AttachJobToRequisition)

		// Jobs
		r.Get("/api/jobs", srv.ListJobs)
		r.Post("/api/jobs", srv.CreateJob)
		r.Get("/api/jobs/{jobId}", srv.GetJob)
		r.Put("/api/jobs/{jobId}", srv.UpdateJob)
		r.Get("/api/jobs/{jobId}/pipeline", srv.GetPipeline)

		// Applications
		r.Get("/api/applications/{appId}", srv.GetApplication)
		r.Put("/api/applications/{appId}/stage", srv.UpdateStage)
		r.Post("/api/applications/{appId}/notes", srv.AddNote)

		// Candidates
		r.Get("/api/candidates", srv.ListCandidates)
		r.Get("/api/candidates/{candidateId}", srv.GetCandidate)
	})

	// Serve React SPA for non-API routes
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("failed to get static sub-fs: %v", err)
	}
	fileServer := http.FileServer(http.FS(staticFS))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		// If it looks like a file request (has extension), try serving it
		if strings.Contains(r.URL.Path, ".") {
			fileServer.ServeHTTP(w, r)
			return
		}
		// Otherwise serve index.html for client-side routing
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	httpSrv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("TempleATS server starting on :%s", port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-done
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("server stopped")
}
