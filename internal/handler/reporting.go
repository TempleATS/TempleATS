package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

// ReqReport returns reporting metrics for a requisition.
func (s *Server) ReqReport(w http.ResponseWriter, r *http.Request) {
	reqID := chi.URLParam(r, "reqId")
	orgID := mw.GetOrgID(r.Context())
	ctx := r.Context()

	req, err := s.Queries.GetRequisitionByID(ctx, db.GetRequisitionByIDParams{
		ID:             reqID,
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "requisition not found"})
		return
	}

	reqIDText := pgtype.Text{String: reqID, Valid: true}

	// Funnel counts
	funnelRows, err := s.Queries.FunnelByRequisition(ctx, reqIDText)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get funnel"})
		return
	}

	funnel := map[string]int32{
		"applied": 0, "screening": 0, "interview": 0,
		"offer": 0, "hired": 0, "rejected": 0,
	}
	for _, row := range funnelRows {
		funnel[row.Stage] = row.Count
	}

	// Rejections breakdown
	rejRows, err := s.Queries.RejectionsByRequisition(ctx, reqIDText)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get rejections"})
		return
	}

	rejByReason := map[string]int32{}
	var rejTotal int32
	for _, row := range rejRows {
		if row.RejectionReason.Valid {
			rejByReason[row.RejectionReason.String] = row.Count
			rejTotal += row.Count
		}
	}

	// Per-job breakdown
	byJobRows, err := s.Queries.ByJobBreakdown(ctx, reqIDText)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get job breakdown"})
		return
	}

	// Avg time in stage
	timeRows, err := s.Queries.AvgTimeInStageByRequisition(ctx, reqIDText)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get time metrics"})
		return
	}

	avgDaysInStage := map[string]float64{}
	for _, row := range timeRows {
		if row.Stage.Valid {
			avgDaysInStage[row.Stage.String] = row.AvgDays
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"requisition": req,
		"funnel":      funnel,
		"timeToHire": map[string]interface{}{
			"avgDaysInStage": avgDaysInStage,
		},
		"rejections": map[string]interface{}{
			"total":    rejTotal,
			"byReason": rejByReason,
		},
		"byJob": byJobRows,
		"fillProgress": map[string]interface{}{
			"hired":  funnel["hired"],
			"target": req.TargetHires,
		},
	})
}
