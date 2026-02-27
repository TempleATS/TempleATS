package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

type createJobRequest struct {
	Title            string  `json:"title"`
	CompanyBlurb     string  `json:"companyBlurb"`
	TeamDetails      string  `json:"teamDetails"`
	Responsibilities string  `json:"responsibilities"`
	Qualifications   string  `json:"qualifications"`
	ClosingStatement string  `json:"closingStatement"`
	Location         *string `json:"location"`
	Department       *string `json:"department"`
	Salary           *string `json:"salary"`
	Status           string  `json:"status"`
	RequisitionID    *string `json:"requisitionId"`
}

type updateJobRequest struct {
	Title            string  `json:"title"`
	CompanyBlurb     string  `json:"companyBlurb"`
	TeamDetails      string  `json:"teamDetails"`
	Responsibilities string  `json:"responsibilities"`
	Qualifications   string  `json:"qualifications"`
	ClosingStatement string  `json:"closingStatement"`
	Location         *string `json:"location"`
	Department       *string `json:"department"`
	Salary           *string `json:"salary"`
	Status           string  `json:"status"`
	RequisitionID    *string `json:"requisitionId"`
}

// ListJobs returns jobs for the current org, filtered by role.
// Supports ?q= query parameter for full-text + fuzzy search.
func (s *Server) ListJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	userID := mw.GetUserID(ctx)
	role := mw.GetRole(ctx)
	query := r.URL.Query().Get("q")

	// Search mode
	if query != "" {
		switch role {
		case "hiring_manager":
			results, err := db.SearchJobsByHiringManager(ctx, s.Pool, orgID, userID, query)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search failed"})
				return
			}
			writeJSON(w, http.StatusOK, results)
		case "interviewer":
			// Interviewers fall back to unfiltered list (no search scoping)
			jobs, err := s.Queries.ListJobsByInterviewer(ctx, db.ListJobsByInterviewerParams{
				OrganizationID: orgID,
				InterviewerID:  userID,
			})
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list jobs"})
				return
			}
			writeJSON(w, http.StatusOK, jobs)
		default:
			results, err := db.SearchJobs(ctx, s.Pool, orgID, query)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search failed"})
				return
			}
			writeJSON(w, http.StatusOK, results)
		}
		return
	}

	switch role {
	case "hiring_manager":
		jobs, err := s.Queries.ListJobsByHiringManager(ctx, db.ListJobsByHiringManagerParams{
			OrganizationID: orgID,
			HiringManagerID: pgtype.Text{String: userID, Valid: true},
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list jobs"})
			return
		}
		writeJSON(w, http.StatusOK, jobs)
	case "interviewer":
		jobs, err := s.Queries.ListJobsByInterviewer(ctx, db.ListJobsByInterviewerParams{
			OrganizationID: orgID,
			InterviewerID:  userID,
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list jobs"})
			return
		}
		writeJSON(w, http.StatusOK, jobs)
	default: // super_admin, admin
		jobs, err := s.Queries.ListJobsByOrg(ctx, orgID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list jobs"})
			return
		}
		writeJSON(w, http.StatusOK, jobs)
	}
}

// CreateJob creates a new job posting.
func (s *Server) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req createJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)

	status := req.Status
	if status == "" {
		status = "draft"
	}

	var location, department, salary, reqID pgtype.Text
	if req.Location != nil {
		location = pgtype.Text{String: *req.Location, Valid: true}
	}
	if req.Department != nil {
		department = pgtype.Text{String: *req.Department, Valid: true}
	}
	if req.Salary != nil {
		salary = pgtype.Text{String: *req.Salary, Valid: true}
	}
	if req.RequisitionID != nil {
		reqID = pgtype.Text{String: *req.RequisitionID, Valid: true}
	}

	job, err := s.Queries.CreateJob(ctx, db.CreateJobParams{
		Title:            req.Title,
		CompanyBlurb:     req.CompanyBlurb,
		TeamDetails:      req.TeamDetails,
		Responsibilities: req.Responsibilities,
		Qualifications:   req.Qualifications,
		ClosingStatement: req.ClosingStatement,
		Location:         location,
		Department:       department,
		Salary:           salary,
		Status:           status,
		RequisitionID:    reqID,
		OrganizationID:   orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create job"})
		return
	}

	go db.InsertAuditLog(context.Background(), s.Pool, orgID, mw.GetUserID(ctx), "create", "job", job.ID, map[string]string{"title": job.Title})
	writeJSON(w, http.StatusCreated, job)
}

// GetJob returns a single job.
func (s *Server) GetJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	jobID := chi.URLParam(r, "jobId")

	job, err := s.Queries.GetJobByID(ctx, db.GetJobByIDParams{
		ID:             jobID,
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}

	writeJSON(w, http.StatusOK, job)
}

// UpdateJob updates a job posting.
func (s *Server) UpdateJob(w http.ResponseWriter, r *http.Request) {
	var req updateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	jobID := chi.URLParam(r, "jobId")

	status := req.Status
	if status == "" {
		status = "draft"
	}

	var location, department, salary, reqID pgtype.Text
	if req.Location != nil {
		location = pgtype.Text{String: *req.Location, Valid: true}
	}
	if req.Department != nil {
		department = pgtype.Text{String: *req.Department, Valid: true}
	}
	if req.Salary != nil {
		salary = pgtype.Text{String: *req.Salary, Valid: true}
	}
	if req.RequisitionID != nil {
		reqID = pgtype.Text{String: *req.RequisitionID, Valid: true}
	}

	job, err := s.Queries.UpdateJob(ctx, db.UpdateJobParams{
		ID:               jobID,
		Title:            req.Title,
		CompanyBlurb:     req.CompanyBlurb,
		TeamDetails:      req.TeamDetails,
		Responsibilities: req.Responsibilities,
		Qualifications:   req.Qualifications,
		ClosingStatement: req.ClosingStatement,
		Location:         location,
		Department:       department,
		Salary:           salary,
		Status:           status,
		RequisitionID:    reqID,
		OrganizationID:   orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}

	go db.InsertAuditLog(context.Background(), s.Pool, orgID, mw.GetUserID(ctx), "update", "job", jobID, map[string]string{"title": req.Title})
	writeJSON(w, http.StatusOK, job)
}
