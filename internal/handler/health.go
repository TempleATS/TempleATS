package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/temple-ats/TempleATS/internal/db"
)

// Server holds shared dependencies for all handlers.
type Server struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
}

// NewServer creates a new Server with the given dependencies.
func NewServer(pool *pgxpool.Pool) *Server {
	return &Server{
		Pool:    pool,
		Queries: db.New(pool),
	}
}

// Health returns a simple health check response.
func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
