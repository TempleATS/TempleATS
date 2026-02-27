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

type createReqRequest struct {
	Title           string  `json:"title"`
	JobCode         *string `json:"jobCode"`
	Level           *string `json:"level"`
	Department      *string `json:"department"`
	TargetHires     *int32  `json:"targetHires"`
	HiringManagerID *string `json:"hiringManagerId"`
	RecruiterID     *string `json:"recruiterId"`
}

type updateReqRequest struct {
	Title           string  `json:"title"`
	JobCode         *string `json:"jobCode"`
	Level           *string `json:"level"`
	Department      *string `json:"department"`
	TargetHires     *int32  `json:"targetHires"`
	Status          string  `json:"status"`
	HiringManagerID *string `json:"hiringManagerId"`
	RecruiterID     *string `json:"recruiterId"`
}

type attachJobRequest struct {
	JobID string `json:"jobId"`
}

// ListRequisitions returns requisitions for the current org, filtered by role.
// Supports ?q= query parameter for full-text + fuzzy search.
func (s *Server) ListRequisitions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	userID := mw.GetUserID(ctx)
	role := mw.GetRole(ctx)
	query := r.URL.Query().Get("q")

	// Interviewers don't see requisitions
	if role == "interviewer" {
		writeJSON(w, http.StatusOK, []interface{}{})
		return
	}

	// Search mode
	if query != "" {
		if role == "hiring_manager" {
			results, err := db.SearchRequisitionsByHiringManager(ctx, s.Pool, orgID, userID, query)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search failed"})
				return
			}
			writeJSON(w, http.StatusOK, results)
			return
		}
		results, err := db.SearchRequisitions(ctx, s.Pool, orgID, query)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search failed"})
			return
		}
		writeJSON(w, http.StatusOK, results)
		return
	}

	if role == "hiring_manager" {
		reqs, err := s.Queries.ListRequisitionsByHiringManager(ctx, db.ListRequisitionsByHiringManagerParams{
			OrganizationID:  orgID,
			HiringManagerID: pgtype.Text{String: userID, Valid: true},
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list requisitions"})
			return
		}
		writeJSON(w, http.StatusOK, reqs)
		return
	}

	// super_admin, admin: see all
	reqs, err := s.Queries.ListRequisitionsByOrg(ctx, orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list requisitions"})
		return
	}

	writeJSON(w, http.StatusOK, reqs)
}

// CreateRequisition creates a new requisition.
func (s *Server) CreateRequisition(w http.ResponseWriter, r *http.Request) {
	var req createReqRequest
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
	userID := mw.GetUserID(ctx)

	targetHires := int32(1)
	if req.TargetHires != nil {
		targetHires = *req.TargetHires
	}

	var jobCode, level, department pgtype.Text
	if req.JobCode != nil {
		jobCode = pgtype.Text{String: *req.JobCode, Valid: true}
	}
	if req.Level != nil {
		level = pgtype.Text{String: *req.Level, Valid: true}
	}
	if req.Department != nil {
		department = pgtype.Text{String: *req.Department, Valid: true}
	}

	hmID := userID
	if req.HiringManagerID != nil && *req.HiringManagerID != "" {
		hmID = *req.HiringManagerID
	}

	var recruiterID pgtype.Text
	if req.RecruiterID != nil && *req.RecruiterID != "" {
		recruiterID = pgtype.Text{String: *req.RecruiterID, Valid: true}
	}

	result, err := s.Queries.CreateRequisition(ctx, db.CreateRequisitionParams{
		Title:            req.Title,
		JobCode:          jobCode,
		Level:            level,
		Department:       department,
		TargetHires:      targetHires,
		HiringManagerID:  pgtype.Text{String: hmID, Valid: true},
		RecruiterID:      recruiterID,
		OrganizationID:   orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create requisition"})
		return
	}

	go db.InsertAuditLog(context.Background(), s.Pool, orgID, userID, "create", "requisition", result.ID, map[string]string{"title": req.Title})
	writeJSON(w, http.StatusCreated, result)
}

// GetRequisition returns a single requisition with its attached jobs.
func (s *Server) GetRequisition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	reqID := chi.URLParam(r, "reqId")

	requisition, err := s.Queries.GetRequisitionByID(ctx, db.GetRequisitionByIDParams{
		ID:             reqID,
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "requisition not found"})
		return
	}

	jobs, err := s.Queries.ListJobsByRequisition(ctx, db.ListJobsByRequisitionParams{
		RequisitionID:  pgtype.Text{String: reqID, Valid: true},
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list jobs"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"requisition": requisition,
		"jobs":        jobs,
	})
}

// UpdateRequisition updates a requisition.
func (s *Server) UpdateRequisition(w http.ResponseWriter, r *http.Request) {
	var req updateReqRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	reqID := chi.URLParam(r, "reqId")

	targetHires := int32(1)
	if req.TargetHires != nil {
		targetHires = *req.TargetHires
	}

	var jobCode, level, department pgtype.Text
	if req.JobCode != nil {
		jobCode = pgtype.Text{String: *req.JobCode, Valid: true}
	}
	if req.Level != nil {
		level = pgtype.Text{String: *req.Level, Valid: true}
	}
	if req.Department != nil {
		department = pgtype.Text{String: *req.Department, Valid: true}
	}

	status := req.Status
	if status == "" {
		status = "open"
	}

	var closedAt pgtype.Timestamptz
	if status == "filled" || status == "cancelled" {
		closedAt = pgtype.Timestamptz{Valid: true}
	}

	// Fetch existing to preserve current values when not provided
	existing, _ := s.Queries.GetRequisitionByID(ctx, db.GetRequisitionByIDParams{
		ID:             reqID,
		OrganizationID: orgID,
	})

	// Resolve hiring manager ID: use provided value, or keep existing
	var hmID pgtype.Text
	if req.HiringManagerID != nil && *req.HiringManagerID != "" {
		hmID = pgtype.Text{String: *req.HiringManagerID, Valid: true}
	} else {
		hmID = existing.HiringManagerID
	}

	// Resolve recruiter ID: use provided value, or keep existing
	var recruiterID pgtype.Text
	if req.RecruiterID != nil && *req.RecruiterID != "" {
		recruiterID = pgtype.Text{String: *req.RecruiterID, Valid: true}
	} else if req.RecruiterID != nil && *req.RecruiterID == "" {
		// Explicitly unset
		recruiterID = pgtype.Text{}
	} else {
		recruiterID = existing.RecruiterID
	}

	result, err := s.Queries.UpdateRequisition(ctx, db.UpdateRequisitionParams{
		ID:              reqID,
		Title:           req.Title,
		JobCode:         jobCode,
		Level:           level,
		Department:      department,
		TargetHires:     targetHires,
		Status:          status,
		HiringManagerID: hmID,
		RecruiterID:     recruiterID,
		ClosedAt:        closedAt,
		OrganizationID:  orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "requisition not found"})
		return
	}

	go db.InsertAuditLog(context.Background(), s.Pool, orgID, mw.GetUserID(ctx), "update", "requisition", reqID, map[string]string{"title": req.Title, "status": status})
	writeJSON(w, http.StatusOK, result)
}

// AttachJobToRequisition links a job to a requisition.
func (s *Server) AttachJobToRequisition(w http.ResponseWriter, r *http.Request) {
	var req attachJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	reqID := chi.URLParam(r, "reqId")

	// Verify the requisition exists
	_, err := s.Queries.GetRequisitionByID(ctx, db.GetRequisitionByIDParams{
		ID:             reqID,
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "requisition not found"})
		return
	}

	// Update the job's requisition_id
	job, err := s.Queries.GetJobByID(ctx, db.GetJobByIDParams{
		ID:             req.JobID,
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}

	result, err := s.Queries.UpdateJob(ctx, db.UpdateJobParams{
		ID:               job.ID,
		Title:            job.Title,
		CompanyBlurb:     job.CompanyBlurb,
		TeamDetails:      job.TeamDetails,
		Responsibilities: job.Responsibilities,
		Qualifications:   job.Qualifications,
		ClosingStatement: job.ClosingStatement,
		Location:         job.Location,
		Department:       job.Department,
		Salary:           job.Salary,
		Status:           job.Status,
		RequisitionID:    pgtype.Text{String: reqID, Valid: true},
		OrganizationID:   orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to attach job"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// DeleteRequisition deletes a requisition. Jobs linked to it are unlinked first.
func (s *Server) DeleteRequisition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	reqID := chi.URLParam(r, "reqId")

	// Unlink any jobs attached to this req
	jobs, err := s.Queries.ListJobsByRequisition(ctx, db.ListJobsByRequisitionParams{
		RequisitionID:  pgtype.Text{String: reqID, Valid: true},
		OrganizationID: orgID,
	})
	if err == nil {
		for _, job := range jobs {
			s.Queries.UpdateJob(ctx, db.UpdateJobParams{
				ID:               job.ID,
				Title:            job.Title,
				CompanyBlurb:     job.CompanyBlurb,
				TeamDetails:      job.TeamDetails,
				Responsibilities: job.Responsibilities,
				Qualifications:   job.Qualifications,
				ClosingStatement: job.ClosingStatement,
				Location:         job.Location,
				Department:       job.Department,
				Salary:           job.Salary,
				Status:           job.Status,
				RequisitionID:    pgtype.Text{}, // unlink
				OrganizationID:   orgID,
			})
		}
	}

	err = s.Queries.DeleteRequisition(ctx, db.DeleteRequisitionParams{
		ID:             reqID,
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "requisition not found"})
		return
	}

	go db.InsertAuditLog(context.Background(), s.Pool, orgID, mw.GetUserID(ctx), "delete", "requisition", reqID, nil)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
