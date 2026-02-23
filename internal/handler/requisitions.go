package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

type createReqRequest struct {
	Title      string  `json:"title"`
	Level      *string `json:"level"`
	Department *string `json:"department"`
	TargetHires *int32 `json:"targetHires"`
}

type updateReqRequest struct {
	Title        string  `json:"title"`
	Level        *string `json:"level"`
	Department   *string `json:"department"`
	TargetHires  *int32  `json:"targetHires"`
	Status       string  `json:"status"`
}

type attachJobRequest struct {
	JobID string `json:"jobId"`
}

// ListRequisitions returns all requisitions for the current org.
func (s *Server) ListRequisitions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)

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

	var level, department pgtype.Text
	if req.Level != nil {
		level = pgtype.Text{String: *req.Level, Valid: true}
	}
	if req.Department != nil {
		department = pgtype.Text{String: *req.Department, Valid: true}
	}

	result, err := s.Queries.CreateRequisition(ctx, db.CreateRequisitionParams{
		Title:            req.Title,
		Level:            level,
		Department:       department,
		TargetHires:      targetHires,
		HiringManagerID:  pgtype.Text{String: userID, Valid: true},
		OrganizationID:   orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create requisition"})
		return
	}

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

	var level, department pgtype.Text
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

	result, err := s.Queries.UpdateRequisition(ctx, db.UpdateRequisitionParams{
		ID:              reqID,
		Title:           req.Title,
		Level:           level,
		Department:      department,
		TargetHires:     targetHires,
		Status:          status,
		HiringManagerID: pgtype.Text{}, // keep existing
		ClosedAt:        closedAt,
		OrganizationID:  orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "requisition not found"})
		return
	}

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
		ID:             job.ID,
		Title:          job.Title,
		Description:    job.Description,
		Location:       job.Location,
		Department:     job.Department,
		Salary:         job.Salary,
		Status:         job.Status,
		RequisitionID:  pgtype.Text{String: reqID, Valid: true},
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to attach job"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}
