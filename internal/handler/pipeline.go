package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
	"github.com/temple-ats/TempleATS/internal/email"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

var validStages = map[string]bool{
	"applied":          true,
	"hr_screen":        true,
	"hm_review":        true,
	"first_interview":  true,
	"final_interview":  true,
	"offer":            true,
	"rejected":         true,
}

// GetPipeline returns applications grouped by stage for a job.
func (s *Server) GetPipeline(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobId")
	orgID := mw.GetOrgID(r.Context())
	userID := mw.GetUserID(r.Context())
	role := mw.GetRole(r.Context())
	ctx := r.Context()

	// Verify job belongs to this org
	job, err := s.Queries.GetJobByID(ctx, db.GetJobByIDParams{ID: jobID, OrganizationID: orgID})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}

	// HM: verify they own this job's requisition
	if role == "hiring_manager" {
		if !job.RequisitionID.Valid {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		}
		req, err := s.Queries.GetRequisitionByID(ctx, db.GetRequisitionByIDParams{
			ID: job.RequisitionID.String, OrganizationID: orgID,
		})
		if err != nil || !req.HiringManagerID.Valid || req.HiringManagerID.String != userID {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		}
	}

	// Interviewer: only show their assigned applications
	if role == "interviewer" {
		apps, err := s.Queries.ListApplicationsByJobForInterviewer(ctx, db.ListApplicationsByJobForInterviewerParams{
			JobID:         job.ID,
			InterviewerID: userID,
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list applications"})
			return
		}
		stages := map[string][]db.ListApplicationsByJobForInterviewerRow{
			"applied": {}, "hr_screen": {}, "hm_review": {}, "first_interview": {}, "final_interview": {},
			"offer": {}, "rejected": {},
		}
		for _, app := range apps {
			stages[app.Stage] = append(stages[app.Stage], app)
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"job": job, "stages": stages})
		return
	}

	apps, err := s.Queries.ListApplicationsByJob(ctx, job.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list applications"})
		return
	}

	// Group by stage
	stages := map[string][]db.ListApplicationsByJobRow{
		"applied": {}, "hm_review": {}, "first_interview": {}, "final_interview": {},
		"offer": {}, "rejected": {},
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
// Allowed: super_admin, admin, hiring_manager (for their jobs only).
// Route-level middleware already blocks interviewer.
func (s *Server) UpdateStage(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	orgID := mw.GetOrgID(r.Context())
	userID := mw.GetUserID(r.Context())
	role := mw.GetRole(r.Context())
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

	// HM: verify they own this job
	if role == "hiring_manager" {
		if !job.RequisitionID.Valid {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		}
		reqObj, err := s.Queries.GetRequisitionByID(ctx, db.GetRequisitionByIDParams{
			ID: job.RequisitionID.String, OrganizationID: orgID,
		})
		if err != nil || !reqObj.HiringManagerID.Valid || reqObj.HiringManagerID.String != userID {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		}
	}

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

	go db.InsertAuditLog(context.Background(), s.Pool, orgID, userID, "stage_change", "application", appID, map[string]string{"from": app.Stage, "to": req.Stage})

	// Send stage change email to candidate if template exists and is enabled
	go func() {
		bgCtx := context.Background()
		tmpl, err := s.Queries.GetEmailTemplateByStage(bgCtx, db.GetEmailTemplateByStageParams{
			OrganizationID: orgID,
			Stage:          req.Stage,
		})
		if err != nil || !tmpl.Enabled {
			return
		}

		appDetails, err := s.Queries.GetApplicationWithDetails(bgCtx, appID)
		if err != nil {
			log.Printf("[stage-email] failed to get app details: %v", err)
			return
		}

		data := map[string]string{
			"CandidateName": appDetails.CandidateName,
			"JobTitle":      appDetails.JobTitle,
			"CompanyName":   appDetails.OrgName,
			"Stage":         req.Stage,
		}
		subject, _ := email.RenderTemplate(tmpl.Subject, data)
		body, _ := email.RenderTemplate(tmpl.Body, data)

		s.Email.SendAsync(orgID, email.SendParams{
			To:            appDetails.CandidateEmail,
			RecipientName: appDetails.CandidateName,
			Subject:       subject,
			Body:          body,
			Type:          "stage_change",
			ApplicationID: appID,
			TriggeredByID: userID,
		})
	}()

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

	notes, err := s.Queries.ListNotesByApplication(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list notes"})
		return
	}

	feedback, err := s.Queries.ListFeedbackByApplication(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list feedback"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"application": app,
		"transitions": transitions,
		"notes":       notes,
		"feedback":    feedback,
	})
}
