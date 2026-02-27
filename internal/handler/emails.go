package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
	"github.com/temple-ats/TempleATS/internal/email"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

type sendEmailRequest struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// SendCandidateEmail sends a free-form email to a candidate.
func (s *Server) SendCandidateEmail(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	orgID := mw.GetOrgID(r.Context())
	userID := mw.GetUserID(r.Context())
	ctx := r.Context()

	var req sendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Subject == "" || req.Body == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "subject and body are required"})
		return
	}

	// Verify application belongs to this org
	appDetails, err := s.Queries.GetApplicationWithDetails(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}
	_, err = s.Queries.GetJobByID(ctx, db.GetJobByIDParams{ID: appDetails.JobID, OrganizationID: orgID})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}

	config := s.Email.GetConfig(ctx, orgID)
	if config == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "SMTP not configured. Please configure email settings first."})
		return
	}

	// Send synchronously so user gets immediate feedback
	if err := email.SendEmail(config, appDetails.CandidateEmail, req.Subject, req.Body); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to send email: " + err.Error()})
		return
	}

	// Record notification for audit trail
	s.Queries.CreateNotification(ctx, db.CreateNotificationParams{
		OrganizationID: orgID,
		Type:           "freeform",
		RecipientEmail: appDetails.CandidateEmail,
		RecipientName:  appDetails.CandidateName,
		Subject:        req.Subject,
		Body:           req.Body,
		Status:         "sent",
		ErrorMessage:   pgtype.Text{},
		ApplicationID:  pgtype.Text{String: appID, Valid: true},
		NoteID:         pgtype.Text{},
		TriggeredByID:  pgtype.Text{String: userID, Valid: true},
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

// ListNotifications returns the email audit trail for an application.
func (s *Server) ListNotifications(w http.ResponseWriter, r *http.Request) {
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

	notifications, err := s.Queries.ListNotificationsByApplication(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list notifications"})
		return
	}

	writeJSON(w, http.StatusOK, notifications)
}
