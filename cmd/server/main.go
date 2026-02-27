package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/temple-ats/TempleATS/internal/auth"
	"github.com/temple-ats/TempleATS/internal/handler"
	"github.com/temple-ats/TempleATS/internal/migrate"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
	"github.com/temple-ats/TempleATS/internal/storage"
	"github.com/temple-ats/TempleATS/migrations"
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

	// Run database migrations
	if err := migrate.Run(ctx, pool, migrations.Files, "."); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Initialize storage (local or S3)
	var store storage.Storage
	switch os.Getenv("STORAGE_TYPE") {
	case "s3":
		s3Store, err := storage.NewS3(ctx, storage.S3Config{
			Bucket:    os.Getenv("S3_BUCKET"),
			Region:    os.Getenv("S3_REGION"),
			Endpoint:  os.Getenv("S3_ENDPOINT"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
			KeyPrefix: os.Getenv("S3_KEY_PREFIX"),
		})
		if err != nil {
			log.Fatalf("failed to initialize S3 storage: %v", err)
		}
		store = s3Store
	default:
		uploadDir := os.Getenv("UPLOAD_DIR")
		if uploadDir == "" {
			uploadDir = "./uploads"
		}
		localStore, err := storage.NewLocal(uploadDir, "/uploads/")
		if err != nil {
			log.Fatalf("failed to initialize local storage: %v", err)
		}
		store = localStore
	}

	// Initialize OIDC for SSO (nil if SSO_ISSUER not set)
	oidcCfg, err := auth.NewOIDCConfig(ctx)
	if err != nil {
		log.Fatalf("failed to initialize OIDC: %v", err)
	}

	srv := handler.NewServer(pool, store)
	srv.OIDC = oidcCfg

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)

	// Rate limiting
	rps := 10.0
	burst := 20
	if v := os.Getenv("RATE_LIMIT_RPS"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			rps = f
		}
	}
	if v := os.Getenv("RATE_LIMIT_BURST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			burst = n
		}
	}
	r.Use(mw.RateLimit(rps, burst))

	// CORS: read allowed origins from env, fall back to localhost for dev
	var allowedOrigins []string
	if corsEnv := os.Getenv("CORS_ORIGINS"); corsEnv != "" {
		for _, o := range strings.Split(corsEnv, ",") {
			if o = strings.TrimSpace(o); o != "" {
				allowedOrigins = append(allowedOrigins, o)
			}
		}
	} else {
		allowedOrigins = []string{"http://localhost:5173", "http://localhost:8080"}
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/api/health", srv.Health)

	// Auth routes (public, stricter rate limit)
	authLimiter := mw.RateLimit(3, 5)
	r.Group(func(r chi.Router) {
		r.Use(authLimiter)
		r.Post("/api/auth/signup", srv.Signup)
		r.Post("/api/auth/login", srv.Login)
		r.Post("/api/auth/accept-invite", srv.AcceptInvitation)
	})
	r.Post("/api/auth/logout", srv.Logout)

	// SSO routes (public)
	r.Get("/api/auth/sso/enabled", srv.SSOEnabled)
	r.Get("/api/auth/sso/url", srv.SSOAuthURL)
	r.Get("/api/auth/sso/callback", srv.SSOCallback)

	// Public careers routes (no auth)
	r.Get("/api/careers/{orgSlug}", srv.CareersListJobs)
	r.Get("/api/careers/{orgSlug}/jobs/{jobId}", srv.CareersGetJob)
	r.Post("/api/careers/{orgSlug}/jobs/{jobId}/apply", srv.CareersApply)

	// Public scheduling routes (no auth - candidate booking)
	r.Get("/api/schedule/{token}", srv.GetPublicSchedule)
	r.Post("/api/schedule/{token}/confirm", srv.ConfirmSchedule)

	// Google OAuth callback (cookie-based auth, not middleware)
	r.Get("/api/auth/google/callback", srv.GoogleAuthCallback)

	// Protected routes - all authenticated users
	r.Group(func(r chi.Router) {
		r.Use(mw.RequireAuth)
		r.Get("/api/auth/me", srv.Me)

		// Read endpoints (handler-level role filtering)
		r.Get("/api/reqs", srv.ListRequisitions)
		r.Get("/api/reqs/{reqId}", srv.GetRequisition)
		r.Get("/api/jobs", srv.ListJobs)
		r.Get("/api/jobs/{jobId}", srv.GetJob)
		r.Get("/api/jobs/{jobId}/pipeline", srv.GetPipeline)
		r.Get("/api/candidates", srv.ListCandidates)
		r.Get("/api/candidates/{candidateId}", srv.GetCandidate)
		r.Get("/api/candidates/{candidateId}/contacts", srv.ListCandidateContacts)
		r.Get("/api/applications/{appId}", srv.GetApplication)
		r.Post("/api/applications/{appId}/notes", srv.AddNote)
		r.Get("/api/my-interviews", srv.MyInterviews)
		r.Get("/api/applications/{appId}/interviewers", srv.ListInterviewers)
		r.Get("/api/applications/{appId}/feedback", srv.ListFeedback)
		r.Get("/api/applications/{appId}/notifications", srv.ListNotifications)
		r.Get("/api/team", srv.ListTeam)
		r.Get("/api/settings/defaults", srv.GetOrgDefaults)
		r.Get("/api/auth/google/url", srv.GoogleAuthURL)
		r.Get("/api/account/calendar", srv.GetCalendarConnectionHandler)
		r.Delete("/api/account/calendar", srv.DisconnectCalendar)
		r.Get("/api/applications/{appId}/schedules", srv.ListSchedules)
		r.Get("/api/settings/email-templates", srv.GetEmailTemplates)
		r.Get("/api/referrals", srv.ListReferrals)
		r.Post("/api/referrals", srv.CreateReferral)
		r.Post("/api/referrals/link", srv.CreateReferralLink)

		// Write: super_admin + admin + recruiter + hiring_manager
		r.Group(func(r chi.Router) {
			r.Use(mw.RequireRole("super_admin", "admin", "recruiter", "hiring_manager"))
			r.Put("/api/applications/{appId}/stage", srv.UpdateStage)
			r.Post("/api/applications/{appId}/interviewers", srv.AssignInterviewer)
			r.Delete("/api/applications/{appId}/interviewers/{userId}", srv.RemoveInterviewer)
		})

		// Feedback: all roles including interviewer
		r.Group(func(r chi.Router) {
			r.Use(mw.RequireRole("super_admin", "admin", "recruiter", "hiring_manager", "interviewer"))
			r.Post("/api/applications/{appId}/feedback", srv.AddFeedback)
			r.Put("/api/applications/{appId}/feedback/{feedbackId}", srv.UpdateFeedback)
			r.Delete("/api/applications/{appId}/feedback/{feedbackId}", srv.DeleteFeedback)
		})

		// Write: super_admin + admin + recruiter
		r.Group(func(r chi.Router) {
			r.Use(mw.RequireRole("super_admin", "admin", "recruiter"))
			r.Put("/api/jobs/{jobId}", srv.UpdateJob)
			r.Post("/api/candidates", srv.CreateCandidate)
			r.Put("/api/candidates/{candidateId}", srv.UpdateCandidate)
			r.Post("/api/candidates/{candidateId}/contacts", srv.AddCandidateContact)
			r.Delete("/api/candidates/{candidateId}/contacts/{contactId}", srv.DeleteCandidateContact)
			r.Post("/api/candidates/{candidateId}/resume", srv.UploadCandidateResume)
			r.Post("/api/candidates/{candidateId}/applications", srv.AddCandidateToJob)
			r.Post("/api/team/invite", srv.InviteUser)
			r.Get("/api/reqs/{reqId}/report", srv.ReqReport)
			r.Post("/api/reqs/{reqId}/snapshots", srv.CreateSnapshot)
			r.Get("/api/reqs/{reqId}/snapshots", srv.ListSnapshots)
			r.Delete("/api/reqs/{reqId}/snapshots/{snapId}", srv.DeleteSnapshot)
			r.Get("/api/metrics/dashboard", srv.DashboardMetrics)
			r.Post("/api/applications/{appId}/email", srv.SendCandidateEmail)
			r.Post("/api/applications/{appId}/hiring-packet", srv.GenerateHiringPacket)
			r.Post("/api/applications/{appId}/schedule", srv.CreateSchedule)
			r.Post("/api/applications/{appId}/availability", srv.CheckAvailability)
		})

		// Write: super_admin + admin only
		r.Group(func(r chi.Router) {
			r.Use(mw.RequireRole("super_admin", "admin"))
			r.Post("/api/reqs", srv.CreateRequisition)
			r.Put("/api/reqs/{reqId}", srv.UpdateRequisition)
			r.Delete("/api/reqs/{reqId}", srv.DeleteRequisition)
			r.Post("/api/reqs/{reqId}/jobs", srv.AttachJobToRequisition)
			r.Post("/api/jobs", srv.CreateJob)
			r.Put("/api/team/{userId}", srv.UpdateTeamMember)
			r.Delete("/api/team/{userId}", srv.RemoveTeamMember)
		})

		// Write: super_admin only
		r.Group(func(r chi.Router) {
			r.Use(mw.RequireRole("super_admin"))
			r.Put("/api/settings/defaults", srv.UpdateOrgDefaults)
			r.Put("/api/settings/org-name", srv.UpdateOrgName)
			r.Get("/api/settings/smtp", srv.GetSmtpSettings)
			r.Put("/api/settings/smtp", srv.UpdateSmtpSettings)
			r.Post("/api/settings/smtp/test", srv.TestSmtpSettings)
			r.Put("/api/settings/email-templates", srv.UpdateEmailTemplate)
		})
	})

	// Serve uploaded files (local disk or S3 presigned redirect)
	r.Handle("/uploads/*", store.Handler())

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
