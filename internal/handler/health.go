package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/temple-ats/TempleATS/internal/auth"
	"github.com/temple-ats/TempleATS/internal/db"
	"github.com/temple-ats/TempleATS/internal/email"
	"github.com/temple-ats/TempleATS/internal/storage"
)

// Server holds shared dependencies for all handlers.
type Server struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
	Storage storage.Storage
	Email   *email.Service
	OIDC    *auth.OIDCConfig // nil if SSO is disabled
}

// NewServer creates a new Server with the given dependencies.
func NewServer(pool *pgxpool.Pool, store storage.Storage) *Server {
	q := db.New(pool)
	return &Server{
		Pool:    pool,
		Queries: q,
		Storage: store,
		Email:   email.NewService(q),
	}
}

// Health returns a simple health check response.
func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
