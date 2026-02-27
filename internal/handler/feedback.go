package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

var validFeedbackStages = map[string]bool{
	"hr_screen":        true,
	"hm_review":        true,
	"first_interview":  true,
	"final_interview":  true,
}

var validRecommendations = map[string]bool{
	"1": true,
	"2": true,
	"3": true,
	"4": true,
	"":  true,
}

type addFeedbackRequest struct {
	Stage          string `json:"stage"`
	InterviewType  string `json:"interviewType"`
	Recommendation string `json:"recommendation"`
	Content        string `json:"content"`
}

type updateFeedbackRequest struct {
	InterviewType  string `json:"interviewType"`
	Recommendation string `json:"recommendation"`
	Content        string `json:"content"`
}

// AddFeedback adds interview feedback to an application for a specific stage.
func (s *Server) AddFeedback(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	orgID := mw.GetOrgID(r.Context())
	userID := mw.GetUserID(r.Context())
	ctx := r.Context()

	var req addFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "content is required"})
		return
	}
	if !validFeedbackStages[req.Stage] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid stage, must be hr_screen, hm_review, first_interview, or final_interview"})
		return
	}
	if !validRecommendations[req.Recommendation] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid recommendation, must be 1, 2, 3, or 4"})
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

	var interviewType pgtype.Text
	if req.InterviewType != "" {
		interviewType = pgtype.Text{String: req.InterviewType, Valid: true}
	}

	feedback, err := s.Queries.CreateInterviewFeedback(ctx, db.CreateInterviewFeedbackParams{
		ApplicationID:  appID,
		Stage:          req.Stage,
		InterviewType:  interviewType,
		Recommendation: req.Recommendation,
		Content:        req.Content,
		AuthorID:       userID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create feedback"})
		return
	}

	writeJSON(w, http.StatusCreated, feedback)
}

// ListFeedback returns all feedback for an application.
func (s *Server) ListFeedback(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	orgID := mw.GetOrgID(r.Context())
	ctx := r.Context()

	// Verify application belongs to this org
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

	feedback, err := s.Queries.ListFeedbackByApplication(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list feedback"})
		return
	}

	writeJSON(w, http.StatusOK, feedback)
}

// UpdateFeedback updates an existing feedback entry. Only the author can update.
func (s *Server) UpdateFeedback(w http.ResponseWriter, r *http.Request) {
	feedbackID := chi.URLParam(r, "feedbackId")
	userID := mw.GetUserID(r.Context())
	ctx := r.Context()

	var req updateFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "content is required"})
		return
	}
	if !validRecommendations[req.Recommendation] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid recommendation, must be 1, 2, 3, or 4"})
		return
	}

	// Verify ownership
	existing, err := s.Queries.GetInterviewFeedbackByID(ctx, feedbackID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "feedback not found"})
		return
	}
	if existing.AuthorID != userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "can only edit your own feedback"})
		return
	}

	var interviewType pgtype.Text
	if req.InterviewType != "" {
		interviewType = pgtype.Text{String: req.InterviewType, Valid: true}
	}

	updated, err := s.Queries.UpdateInterviewFeedback(ctx, db.UpdateInterviewFeedbackParams{
		ID:             feedbackID,
		Recommendation: req.Recommendation,
		Content:        req.Content,
		InterviewType:  interviewType,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update feedback"})
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// DeleteFeedback deletes a feedback entry. Author or admin+ can delete.
func (s *Server) DeleteFeedback(w http.ResponseWriter, r *http.Request) {
	feedbackID := chi.URLParam(r, "feedbackId")
	userID := mw.GetUserID(r.Context())
	role := mw.GetRole(r.Context())
	ctx := r.Context()

	existing, err := s.Queries.GetInterviewFeedbackByID(ctx, feedbackID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "feedback not found"})
		return
	}

	if existing.AuthorID != userID && role != "admin" && role != "super_admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "can only delete your own feedback"})
		return
	}

	if err := s.Queries.DeleteInterviewFeedback(ctx, feedbackID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete feedback"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
