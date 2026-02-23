package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

var validStages = map[string]bool{
	"applied":   true,
	"screening": true,
	"interview": true,
	"offer":     true,
	"hired":     true,
	"rejected":  true,
}

// GetPipeline returns applications grouped by stage for a job.
func (s *Server) GetPipeline(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobId")
	orgID := mw.GetOrgID(r.Context())
	ctx := r.Context()

	// Verify job belongs to this org
	job, err := s.Queries.GetJobByID(ctx, db.GetJobByIDParams{ID: jobID, OrganizationID: orgID})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}

	apps, err := s.Queries.ListApplicationsByJob(ctx, job.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list applications"})
		return
	}

	// Group by stage
	stages := map[string][]db.ListApplicationsByJobRow{
		"applied":   {},
		"screening": {},
		"interview": {},
		"offer":     {},
		"hired":     {},
		"rejected":  {},
	}
	for _, app := range apps {
		stages[app.Stage] = append(stages[app.Stage], app)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"job":    job,
		"stages": stages,
	})
}

type updateStageRequest struct {
	Stage           string `json:"stage"`
	RejectionReason string `json:"rejectionReason"`
	RejectionNotes  string `json:"rejectionNotes"`
}

// UpdateStage moves an application to a new stage.
func (s *Server) UpdateStage(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	orgID := mw.GetOrgID(r.Context())
	userID := mw.GetUserID(r.Context())
	ctx := r.Context()

	var req updateStageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if !validStages[req.Stage] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid stage"})
		return
	}

	if req.Stage == "rejected" && req.RejectionReason == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "rejection reason is required"})
		return
	}

	// Get current application and verify org ownership through job
	app, err := s.Queries.GetApplicationByID(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}

	job, err := s.Queries.GetJobByID(ctx, db.GetJobByIDParams{ID: app.JobID, OrganizationID: orgID})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}
	_ = job

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)

	// Update stage
	var rejReason, rejNotes pgtype.Text
	if req.RejectionReason != "" {
		rejReason = pgtype.Text{String: req.RejectionReason, Valid: true}
	}
	if req.RejectionNotes != "" {
		rejNotes = pgtype.Text{String: req.RejectionNotes, Valid: true}
	}

	updated, err := qtx.UpdateApplicationStage(ctx, db.UpdateApplicationStageParams{
		ID:              appID,
		Stage:           req.Stage,
		RejectionReason: rejReason,
		RejectionNotes:  rejNotes,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update stage"})
		return
	}

	// Record transition
	_, err = qtx.CreateStageTransition(ctx, db.CreateStageTransitionParams{
		ApplicationID: appID,
		FromStage:     pgtype.Text{String: app.Stage, Valid: true},
		ToStage:       req.Stage,
		MovedByID:     pgtype.Text{String: userID, Valid: true},
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// GetApplication returns application detail with transitions.
func (s *Server) GetApplication(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	orgID := mw.GetOrgID(r.Context())
	ctx := r.Context()

	app, err := s.Queries.GetApplicationWithDetails(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}

	// Verify org ownership through job
	_, err = s.Queries.GetJobByID(ctx, db.GetJobByIDParams{ID: app.JobID, OrganizationID: orgID})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}

	transitions, err := s.Queries.ListTransitionsByApplication(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list transitions"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"application": app,
		"transitions": transitions,
	})
}
