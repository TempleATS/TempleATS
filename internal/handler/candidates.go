package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

// ListCandidates returns candidates for the current org, with optional search.
func (s *Server) ListCandidates(w http.ResponseWriter, r *http.Request) {
	orgID := mw.GetOrgID(r.Context())
	ctx := r.Context()

	q := r.URL.Query().Get("q")
	if q != "" {
		candidates, err := s.Queries.SearchCandidates(ctx, db.SearchCandidatesParams{
			OrganizationID: orgID,
			Column2:        pgtype.Text{String: q, Valid: true},
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to search candidates"})
			return
		}
		writeJSON(w, http.StatusOK, candidates)
		return
	}

	candidates, err := s.Queries.ListCandidatesByOrg(ctx, orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list candidates"})
		return
	}
	writeJSON(w, http.StatusOK, candidates)
}

// GetCandidate returns a candidate with their applications.
func (s *Server) GetCandidate(w http.ResponseWriter, r *http.Request) {
	candID := chi.URLParam(r, "candidateId")
	orgID := mw.GetOrgID(r.Context())
	ctx := r.Context()

	candidate, err := s.Queries.GetCandidateByID(ctx, db.GetCandidateByIDParams{
		ID:             candID,
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "candidate not found"})
		return
	}

	applications, err := s.Queries.ListApplicationsByCandidate(ctx, candID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list applications"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"candidate":    candidate,
		"applications": applications,
	})
}

type addNoteRequest struct {
	Content string `json:"content"`
}

// AddNote adds a note to an application.
func (s *Server) AddNote(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	orgID := mw.GetOrgID(r.Context())
	userID := mw.GetUserID(r.Context())
	ctx := r.Context()

	var req addNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "content is required"})
		return
	}

	// Verify application belongs to this org through job
	app, err := s.Queries.GetApplicationByID(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}
	_, err = s.Queries.GetJobByID(ctx, db.GetJobByIDParams{ID: app.JobID, OrganizationID: orgID})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}

	note, err := s.Queries.CreateNote(ctx, db.CreateNoteParams{
		Content:       req.Content,
		ApplicationID: appID,
		AuthorID:      userID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create note"})
		return
	}

	writeJSON(w, http.StatusCreated, note)
}
