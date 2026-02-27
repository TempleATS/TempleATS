package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
)

type applyRequest struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	ResumeURL      string `json:"resumeUrl"`
	ResumeFilename string `json:"resumeFilename"`
}

// CareersListJobs returns the org info and open jobs for a public careers page.
func (s *Server) CareersListJobs(w http.ResponseWriter, r *http.Request) {
	orgSlug := chi.URLParam(r, "orgSlug")
	ctx := r.Context()

	org, err := s.Queries.GetOrganizationBySlug(ctx, orgSlug)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "organization not found"})
		return
	}

	jobs, err := s.Queries.ListOpenJobsByOrgSlug(ctx, orgSlug)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list jobs"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"organization": map[string]interface{}{
			"name":    org.Name,
			"slug":    org.Slug,
			"logoUrl": org.LogoUrl,
			"website": org.Website,
		},
		"jobs": jobs,
	})
}

// CareersGetJob returns a single open job for the public careers page.
func (s *Server) CareersGetJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobId")
	ctx := r.Context()

	job, err := s.Queries.GetJobPublic(ctx, jobID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}

	writeJSON(w, http.StatusOK, job)
}

var allowedResumeExts = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
}

const maxResumeSize = 10 << 20 // 10 MB

// CareersApply handles a public application submission.
// Accepts either JSON or multipart/form-data (with resume file).
func (s *Server) CareersApply(w http.ResponseWriter, r *http.Request) {
	orgSlug := chi.URLParam(r, "orgSlug")
	jobID := chi.URLParam(r, "jobId")
	ctx := r.Context()

	var req applyRequest

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(maxResumeSize); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large (max 10MB)"})
			return
		}
		req.Name = r.FormValue("name")
		req.Email = r.FormValue("email")
		req.Phone = r.FormValue("phone")

		// Handle resume file
		file, header, err := r.FormFile("resume")
		if err == nil {
			defer file.Close()

			ext := strings.ToLower(filepath.Ext(header.Filename))
			if !allowedResumeExts[ext] {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resume must be PDF, DOC, or DOCX"})
				return
			}

			savedName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
			url, err := s.Storage.Save(r.Context(), savedName, file)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save resume"})
				return
			}

			req.ResumeURL = url
			req.ResumeFilename = header.Filename
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
	}

	if req.Name == "" || req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and email are required"})
		return
	}

	// Get org by slug
	org, err := s.Queries.GetOrganizationBySlug(ctx, orgSlug)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "organization not found"})
		return
	}

	// Verify job exists, is open, and belongs to this org
	job, err := s.Queries.GetJobPublic(ctx, jobID)
	if err != nil || job.OrganizationID != org.ID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)

	// Upsert candidate
	var phone, resumeURL, resumeFilename pgtype.Text
	if req.Phone != "" {
		phone = pgtype.Text{String: req.Phone, Valid: true}
	}
	if req.ResumeURL != "" {
		resumeURL = pgtype.Text{String: req.ResumeURL, Valid: true}
	}
	if req.ResumeFilename != "" {
		resumeFilename = pgtype.Text{String: req.ResumeFilename, Valid: true}
	}

	candidate, err := qtx.UpsertCandidate(ctx, db.UpsertCandidateParams{
		Name:           req.Name,
		Email:          req.Email,
		Phone:          phone,
		ResumeUrl:      resumeURL,
		ResumeFilename: resumeFilename,
		OrganizationID: org.ID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create candidate"})
		return
	}

	// Create application
	application, err := qtx.CreateApplication(ctx, db.CreateApplicationParams{
		CandidateID: candidate.ID,
		JobID:       jobID,
	})
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "you have already applied to this job"})
		return
	}

	// Handle referral token
	if refToken := r.URL.Query().Get("ref"); refToken != "" {
		ref, err := qtx.GetReferralByToken(ctx, refToken)
		if err == nil && ref.JobID == jobID && ref.OrganizationID == org.ID {
			_ = qtx.SetApplicationReferral(ctx, application.ID, ref.ID)
			_ = qtx.UpdateReferralApplication(ctx, db.UpdateReferralApplicationParams{
				ID:            ref.ID,
				CandidateName: pgtype.Text{String: req.Name, Valid: true},
				CandidateID:   pgtype.Text{String: candidate.ID, Valid: true},
				ApplicationID: pgtype.Text{String: application.ID, Valid: true},
			})
		}
	}

	// Record initial stage transition
	_, err = qtx.CreateStageTransition(ctx, db.CreateStageTransitionParams{
		ApplicationID: application.ID,
		FromStage:     pgtype.Text{},
		ToStage:       "applied",
		MovedByID:     pgtype.Text{},
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"applicationId": application.ID,
		"candidateId":   candidate.ID,
		"message":       "Application submitted successfully",
	})
}
