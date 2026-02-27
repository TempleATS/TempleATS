package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

// ListReferrals returns referrals for the org. Recruiter+ sees all; others see only their own.
func (s *Server) ListReferrals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	userID := mw.GetUserID(ctx)
	role := mw.GetRole(ctx)

	var items []db.ListReferralsByOrgRow
	var err error

	switch role {
	case "super_admin", "admin", "recruiter":
		items, err = s.Queries.ListReferralsByOrg(ctx, orgID)
	default:
		items, err = s.Queries.ListReferralsByUser(ctx, db.ListReferralsByUserParams{
			ReferrerID:     userID,
			OrganizationID: orgID,
		})
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list referrals"})
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// CreateReferral handles a direct referral: team member refers a candidate to a job.
func (s *Server) CreateReferral(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	userID := mw.GetUserID(ctx)

	var name, email, phone, jobID string
	var resumeURL, resumeFilename string

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(maxResumeSize); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large (max 10MB)"})
			return
		}
		name = r.FormValue("name")
		email = r.FormValue("email")
		phone = r.FormValue("phone")
		jobID = r.FormValue("jobId")

		file, header, err := r.FormFile("resume")
		if err == nil {
			defer file.Close()
			ext := strings.ToLower(filepath.Ext(header.Filename))
			if !allowedResumeExts[ext] {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resume must be PDF, DOC, or DOCX"})
				return
			}
			savedName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
			url, err := s.Storage.Save(ctx, savedName, file)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save resume"})
				return
			}
			resumeURL = url
			resumeFilename = header.Filename
		}
	} else {
		var req struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Phone string `json:"phone"`
			JobID string `json:"jobId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		name = req.Name
		email = req.Email
		phone = req.Phone
		jobID = req.JobID
	}

	if name == "" || email == "" || jobID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name, email, and jobId are required"})
		return
	}

	// Verify job belongs to org and is open
	job, err := s.Queries.GetJobByID(ctx, db.GetJobByIDParams{ID: jobID, OrganizationID: orgID})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}
	if job.Status != "open" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "job is not open"})
		return
	}

	// Transaction
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer tx.Rollback(ctx)
	qtx := s.Queries.WithTx(tx)

	// Upsert candidate
	var phonePg, resumeURLPg, resumeFilenamePg pgtype.Text
	if phone != "" {
		phonePg = pgtype.Text{String: phone, Valid: true}
	}
	if resumeURL != "" {
		resumeURLPg = pgtype.Text{String: resumeURL, Valid: true}
	}
	if resumeFilename != "" {
		resumeFilenamePg = pgtype.Text{String: resumeFilename, Valid: true}
	}

	candidate, err := qtx.UpsertCandidate(ctx, db.UpsertCandidateParams{
		Name:           name,
		Email:          email,
		Phone:          phonePg,
		ResumeUrl:      resumeURLPg,
		ResumeFilename: resumeFilenamePg,
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create candidate"})
		return
	}

	// Create referral (without application_id initially)
	referral, err := qtx.CreateReferral(ctx, db.CreateReferralParams{
		ReferrerID:     userID,
		OrganizationID: orgID,
		JobID:          jobID,
		Source:         "direct",
		CandidateName:  pgtype.Text{String: name, Valid: true},
		CandidateID:    pgtype.Text{String: candidate.ID, Valid: true},
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create referral"})
		return
	}

	// Create application with referral link
	application, err := qtx.CreateApplicationWithReferral(ctx, db.CreateApplicationWithReferralParams{
		CandidateID: candidate.ID,
		JobID:       jobID,
		ReferralID:  referral.ID,
	})
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "this candidate already has an application for this job"})
		return
	}

	// Update referral with the application ID
	_ = qtx.UpdateReferralApplication(ctx, db.UpdateReferralApplicationParams{
		ID:            referral.ID,
		CandidateName: pgtype.Text{String: name, Valid: true},
		CandidateID:   pgtype.Text{String: candidate.ID, Valid: true},
		ApplicationID: pgtype.Text{String: application.ID, Valid: true},
	})

	// Record stage transition
	_, _ = qtx.CreateStageTransition(ctx, db.CreateStageTransitionParams{
		ApplicationID: application.ID,
		FromStage:     pgtype.Text{},
		ToStage:       "applied",
		MovedByID:     pgtype.Text{String: userID, Valid: true},
	})

	if err := tx.Commit(ctx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"referralId":    referral.ID,
		"applicationId": application.ID,
		"candidateId":   candidate.ID,
	})
}

// CreateReferralLink generates a shareable referral link for a job.
func (s *Server) CreateReferralLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	userID := mw.GetUserID(ctx)

	var req struct {
		JobID string `json:"jobId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.JobID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "jobId is required"})
		return
	}

	// Verify job
	job, err := s.Queries.GetJobByID(ctx, db.GetJobByIDParams{ID: req.JobID, OrganizationID: orgID})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}
	if job.Status != "open" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "job is not open"})
		return
	}

	referral, err := s.Queries.CreateReferral(ctx, db.CreateReferralParams{
		ReferrerID:     userID,
		OrganizationID: orgID,
		JobID:          req.JobID,
		Source:         "link",
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create referral link"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"referralId": referral.ID,
		"token":      referral.Token,
	})
}
