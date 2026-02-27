package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
	"github.com/temple-ats/TempleATS/internal/email"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

// ListCandidates returns candidates for the current org, filtered by role.
func (s *Server) ListCandidates(w http.ResponseWriter, r *http.Request) {
	orgID := mw.GetOrgID(r.Context())
	userID := mw.GetUserID(r.Context())
	role := mw.GetRole(r.Context())
	ctx := r.Context()

	// For search, only super_admin and admin can search all candidates
	q := r.URL.Query().Get("q")
	if q != "" && (role == "super_admin" || role == "admin" || role == "recruiter") {
		candidates, err := s.Queries.SearchCandidatesWithAllApps(ctx, db.SearchCandidatesWithAllAppsParams{
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

	switch role {
	case "hiring_manager":
		candidates, err := s.Queries.ListCandidatesWithLatestAppByHM(ctx, db.ListCandidatesWithLatestAppByHMParams{
			OrganizationID:  orgID,
			HiringManagerID: pgtype.Text{String: userID, Valid: true},
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list candidates"})
			return
		}
		writeJSON(w, http.StatusOK, candidates)
	case "interviewer":
		candidates, err := s.Queries.ListCandidatesWithLatestAppByInterviewer(ctx, db.ListCandidatesWithLatestAppByInterviewerParams{
			OrganizationID: orgID,
			InterviewerID:  userID,
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list candidates"})
			return
		}
		writeJSON(w, http.StatusOK, candidates)
	default: // super_admin, admin
		candidates, err := s.Queries.ListCandidatesWithAllApps(ctx, orgID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list candidates"})
			return
		}
		writeJSON(w, http.StatusOK, candidates)
	}
}

// CreateCandidate creates a new candidate with optional resume and optional job application.
func (s *Server) CreateCandidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	userID := mw.GetUserID(ctx)

	var name, emailAddr, phone, jobID string
	var resumeURL, resumeFilename string

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(maxResumeSize); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large (max 10MB)"})
			return
		}
		name = r.FormValue("name")
		emailAddr = r.FormValue("email")
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
			destPath := filepath.Join(s.UploadDir, savedName)
			dst, err := os.Create(destPath)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save resume"})
				return
			}
			defer dst.Close()
			if _, err := io.Copy(dst, file); err != nil {
				os.Remove(destPath)
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save resume"})
				return
			}
			resumeURL = "/uploads/" + savedName
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
		emailAddr = req.Email
		phone = req.Phone
		jobID = req.JobID
	}

	if name == "" || emailAddr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and email are required"})
		return
	}

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

	candidate, err := s.Queries.UpsertCandidate(ctx, db.UpsertCandidateParams{
		Name:           name,
		Email:          emailAddr,
		Phone:          phonePg,
		ResumeUrl:      resumeURLPg,
		ResumeFilename: resumeFilenamePg,
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create candidate"})
		return
	}

	var applicationID string
	if jobID != "" {
		job, err := s.Queries.GetJobByID(ctx, db.GetJobByIDParams{ID: jobID, OrganizationID: orgID})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "job not found"})
			return
		}
		if job.Status != "open" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "job is not open"})
			return
		}

		app, err := s.Queries.CreateApplication(ctx, db.CreateApplicationParams{
			CandidateID: candidate.ID,
			JobID:       jobID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "this candidate already has an application for this job"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create application"})
			return
		}
		applicationID = app.ID

		_, _ = s.Queries.CreateStageTransition(ctx, db.CreateStageTransitionParams{
			ApplicationID: app.ID,
			FromStage:     pgtype.Text{},
			ToStage:       "applied",
			MovedByID:     pgtype.Text{String: userID, Valid: true},
		})
	}

	resp := map[string]string{"candidateId": candidate.ID}
	if applicationID != "" {
		resp["applicationId"] = applicationID
	}
	writeJSON(w, http.StatusCreated, resp)
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

type addToJobRequest struct {
	JobID string `json:"jobId"`
}

// AddCandidateToJob creates an application for an existing candidate on a job.
func (s *Server) AddCandidateToJob(w http.ResponseWriter, r *http.Request) {
	candID := chi.URLParam(r, "candidateId")
	orgID := mw.GetOrgID(r.Context())
	ctx := r.Context()

	var req addToJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.JobID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "jobId is required"})
		return
	}

	// Verify candidate belongs to this org
	_, err := s.Queries.GetCandidateByID(ctx, db.GetCandidateByIDParams{
		ID: candID, OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "candidate not found"})
		return
	}

	// Verify job belongs to this org
	_, err = s.Queries.GetJobByID(ctx, db.GetJobByIDParams{
		ID: req.JobID, OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}

	app, err := s.Queries.CreateApplication(ctx, db.CreateApplicationParams{
		CandidateID: candID,
		JobID:       req.JobID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "candidate already applied to this job"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create application"})
		return
	}

	writeJSON(w, http.StatusCreated, app)
}

type updateCandidateRequest struct {
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	LinkedinUrl string `json:"linkedinUrl"`
}

// UpdateCandidate updates a candidate's contact info.
func (s *Server) UpdateCandidate(w http.ResponseWriter, r *http.Request) {
	candID := chi.URLParam(r, "candidateId")
	orgID := mw.GetOrgID(r.Context())
	ctx := r.Context()

	var req updateCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email is required"})
		return
	}

	candidate, err := s.Queries.UpdateCandidateContact(ctx, db.UpdateCandidateContactParams{
		ID:             candID,
		OrganizationID: orgID,
		Email:          req.Email,
		Phone:          pgtype.Text{String: req.Phone, Valid: req.Phone != ""},
		LinkedinUrl:    pgtype.Text{String: req.LinkedinUrl, Valid: req.LinkedinUrl != ""},
	})
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "candidate not found"})
			return
		}
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "a candidate with that email already exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update candidate"})
		return
	}

	writeJSON(w, http.StatusOK, candidate)
}

// ListCandidateContacts returns additional contacts for a candidate.
func (s *Server) ListCandidateContacts(w http.ResponseWriter, r *http.Request) {
	candID := chi.URLParam(r, "candidateId")
	orgID := mw.GetOrgID(r.Context())
	ctx := r.Context()

	// Verify candidate belongs to this org
	_, err := s.Queries.GetCandidateByID(ctx, db.GetCandidateByIDParams{
		ID: candID, OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "candidate not found"})
		return
	}

	contacts, err := s.Queries.ListContactsByCandidate(ctx, candID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list contacts"})
		return
	}
	writeJSON(w, http.StatusOK, contacts)
}

var validCategories = map[string]bool{"email": true, "phone": true, "online_presence": true}

type addContactRequest struct {
	Category string `json:"category"`
	Label    string `json:"label"`
	Value    string `json:"value"`
}

// AddCandidateContact adds a contact entry for a candidate.
func (s *Server) AddCandidateContact(w http.ResponseWriter, r *http.Request) {
	candID := chi.URLParam(r, "candidateId")
	orgID := mw.GetOrgID(r.Context())
	ctx := r.Context()

	var req addContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if !validCategories[req.Category] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "category must be email, phone, or online_presence"})
		return
	}
	if req.Label == "" || req.Value == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "label and value are required"})
		return
	}

	// Verify candidate belongs to this org
	_, err := s.Queries.GetCandidateByID(ctx, db.GetCandidateByIDParams{
		ID: candID, OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "candidate not found"})
		return
	}

	contact, err := s.Queries.CreateCandidateContact(ctx, db.CreateCandidateContactParams{
		CandidateID: candID,
		Category:    req.Category,
		Label:       req.Label,
		Value:       req.Value,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "this contact already exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add contact"})
		return
	}
	writeJSON(w, http.StatusCreated, contact)
}

// DeleteCandidateContact removes a contact entry.
func (s *Server) DeleteCandidateContact(w http.ResponseWriter, r *http.Request) {
	candID := chi.URLParam(r, "candidateId")
	contactID := chi.URLParam(r, "contactId")
	orgID := mw.GetOrgID(r.Context())
	ctx := r.Context()

	// Verify candidate belongs to this org
	_, err := s.Queries.GetCandidateByID(ctx, db.GetCandidateByIDParams{
		ID: candID, OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "candidate not found"})
		return
	}

	err = s.Queries.DeleteCandidateContact(ctx, db.DeleteCandidateContactParams{
		ID:          contactID,
		CandidateID: candID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete contact"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// UploadCandidateResume handles resume file upload for a candidate.
func (s *Server) UploadCandidateResume(w http.ResponseWriter, r *http.Request) {
	candID := chi.URLParam(r, "candidateId")
	orgID := mw.GetOrgID(r.Context())
	ctx := r.Context()

	if err := r.ParseMultipartForm(maxResumeSize); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large (max 10MB)"})
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resume file is required"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedResumeExts[ext] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resume must be PDF, DOC, or DOCX"})
		return
	}

	savedName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	destPath := filepath.Join(s.UploadDir, savedName)
	dst, err := os.Create(destPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save resume"})
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(destPath)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save resume"})
		return
	}

	resumeURL := "/uploads/" + savedName
	candidate, err := s.Queries.UpdateCandidateResume(ctx, db.UpdateCandidateResumeParams{
		ID:             candID,
		OrganizationID: orgID,
		ResumeUrl:      pgtype.Text{String: resumeURL, Valid: true},
		ResumeFilename: pgtype.Text{String: header.Filename, Valid: true},
	})
	if err != nil {
		os.Remove(destPath)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update candidate"})
		return
	}

	writeJSON(w, http.StatusOK, candidate)
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

	// Detect @mentions and send email notifications
	if strings.Contains(req.Content, "@") {
		go func() {
			bgCtx := context.Background()
			users, err := s.Queries.ListUsersByOrg(bgCtx, orgID)
			if err != nil {
				log.Printf("[mentions] failed to list users: %v", err)
				return
			}
			mentioned := email.ParseMentions(req.Content, users)
			if len(mentioned) == 0 {
				return
			}

			mentioner, _ := s.Queries.GetUserByID(bgCtx, userID)
			mentionerName := "A team member"
			if mentioner.Name != "" {
				mentionerName = mentioner.Name
			}
			appDetails, _ := s.Queries.GetApplicationWithDetails(bgCtx, appID)

			for _, user := range mentioned {
				if user.ID == userID {
					continue
				}
				subject := fmt.Sprintf("%s mentioned you on %s's application", mentionerName, appDetails.CandidateName)
				body := fmt.Sprintf(
					"<p><strong>%s</strong> mentioned you in a note on <strong>%s</strong>'s application for <strong>%s</strong>:</p><blockquote style=\"border-left:3px solid #ccc;padding-left:12px;color:#555;\">%s</blockquote>",
					mentionerName, appDetails.CandidateName, appDetails.JobTitle, req.Content,
				)
				s.Email.SendAsync(orgID, email.SendParams{
					To:            user.Email,
					RecipientName: user.Name,
					Subject:       subject,
					Body:          body,
					Type:          "mention",
					ApplicationID: appID,
					NoteID:        note.ID,
					TriggeredByID: userID,
				})
			}
		}()
	}

	writeJSON(w, http.StatusCreated, note)
}
