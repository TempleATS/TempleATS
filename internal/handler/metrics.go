package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

// CreateSnapshot computes live metrics for a requisition and saves a snapshot.
func (s *Server) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	reqID := chi.URLParam(r, "reqId")
	orgID := mw.GetOrgID(r.Context())
	userID := mw.GetUserID(r.Context())
	ctx := r.Context()

	// Parse optional label
	var body struct {
		Label string `json:"label"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	// Verify req belongs to org
	req, err := s.Queries.GetRequisitionByID(ctx, db.GetRequisitionByIDParams{
		ID: reqID, OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "requisition not found"})
		return
	}

	repo := db.NewMetricSnapshotRepository(s.Pool)

	// Get current funnel counts
	reqIDText := pgtype.Text{String: reqID, Valid: true}
	funnelRows, err := s.Queries.FunnelByRequisition(ctx, reqIDText)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get funnel"})
		return
	}

	funnel := map[string]int{}
	for _, row := range funnelRows {
		funnel[row.Stage] = int(row.Count)
	}

	// Get throughput counts
	tpRows, err := repo.ThroughputByRequisition(ctx, reqID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get throughput"})
		return
	}

	tp := map[string]int{}
	for _, row := range tpRows {
		tp[row.Stage] = row.Count
	}

	// Compute total
	total := 0
	for _, v := range funnel {
		total += v
	}

	// Build snapshot
	snap := &db.MetricSnapshot{
		RequisitionID:  reqID,
		OrganizationID: orgID,
		CreatedByID:    userID,
		FunnelApplied:  funnel["applied"],
		FunnelHRScreen: funnel["hr_screen"],
		FunnelHMReview: funnel["hm_review"],
		FunnelFirstIV:  funnel["first_interview"],
		FunnelFinalIV:  funnel["final_interview"],
		FunnelOffer:    funnel["offer"],
		FunnelHired:    funnel["hired"],
		FunnelRejected: funnel["rejected"],
		TPApplied:      tp["applied"],
		TPFirstIV:      tp["first_interview"],
		TPFinalIV:      tp["final_interview"],
		TPOffer:        tp["offer"],
		TPHired:        tp["hired"],
		TotalApps:      total,
		TargetHires:    int(req.TargetHires),
	}

	if body.Label != "" {
		snap.Label = &body.Label
	}

	// Compute ratios
	if tp["first_interview"] > 0 {
		v := float64(tp["final_interview"]) / float64(tp["first_interview"])
		snap.RatioFirstFinal = &v
	}
	if tp["final_interview"] > 0 {
		v := float64(tp["offer"]) / float64(tp["final_interview"])
		snap.RatioFinalOffer = &v
	}
	if tp["offer"] > 0 {
		v := float64(tp["hired"]) / float64(tp["offer"])
		snap.RatioOfferHired = &v
	}

	if err := repo.InsertSnapshot(ctx, snap); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save snapshot"})
		return
	}

	writeJSON(w, http.StatusCreated, snap)
}

// ListSnapshots returns saved snapshots for a requisition.
func (s *Server) ListSnapshots(w http.ResponseWriter, r *http.Request) {
	reqID := chi.URLParam(r, "reqId")
	ctx := r.Context()

	repo := db.NewMetricSnapshotRepository(s.Pool)
	snapshots, err := repo.ListSnapshotsByReq(ctx, reqID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list snapshots"})
		return
	}

	if snapshots == nil {
		snapshots = []db.MetricSnapshot{}
	}
	writeJSON(w, http.StatusOK, snapshots)
}

// DeleteSnapshot deletes a saved snapshot.
func (s *Server) DeleteSnapshot(w http.ResponseWriter, r *http.Request) {
	snapID := chi.URLParam(r, "snapId")
	ctx := r.Context()

	repo := db.NewMetricSnapshotRepository(s.Pool)
	if err := repo.DeleteSnapshot(ctx, snapID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete snapshot"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// DashboardMetrics returns org-wide metrics with per-recruiter and per-HM stats.
func (s *Server) DashboardMetrics(w http.ResponseWriter, r *http.Request) {
	orgID := mw.GetOrgID(r.Context())
	ctx := r.Context()

	repo := db.NewMetricSnapshotRepository(s.Pool)

	// Per-recruiter stats
	recruiterStats, err := repo.RecruiterStats(ctx, orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get recruiter stats"})
		return
	}
	if recruiterStats == nil {
		recruiterStats = []db.PersonStats{}
	}

	// Per-HM stats
	hmStats, err := repo.HiringManagerStats(ctx, orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get HM stats"})
		return
	}
	if hmStats == nil {
		hmStats = []db.PersonStats{}
	}

	// Org-wide throughput
	orgTP, err := repo.OrgThroughput(ctx, orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get org throughput"})
		return
	}

	// Org-wide totals from recruiter stats
	var totalReqs, openReqs, totalApps, totalHired, totalRejected int
	for _, rs := range recruiterStats {
		totalReqs += rs.TotalReqs
		openReqs += rs.OpenReqs
		totalApps += rs.TotalApps
		totalHired += rs.TotalHired
		totalRejected += rs.TotalRejected
	}

	// Org-wide conversion ratios from throughput
	orgConversions := map[string]interface{}{
		"tp_first_interview": orgTP["first_interview"],
		"tp_final_interview": orgTP["final_interview"],
		"tp_offer":           orgTP["offer"],
		"tp_hired":           orgTP["hired"],
		"tp_applied":         orgTP["applied"],
	}
	if orgTP["first_interview"] > 0 {
		orgConversions["ratio_first_to_final"] = float64(orgTP["final_interview"]) / float64(orgTP["first_interview"])
	}
	if orgTP["final_interview"] > 0 {
		orgConversions["ratio_final_to_offer"] = float64(orgTP["offer"]) / float64(orgTP["final_interview"])
	}
	if orgTP["offer"] > 0 {
		orgConversions["ratio_offer_to_hired"] = float64(orgTP["hired"]) / float64(orgTP["offer"])
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_reqs":       totalReqs,
		"open_reqs":        openReqs,
		"total_applications": totalApps,
		"total_hired":       totalHired,
		"total_rejected":    totalRejected,
		"org_conversions":   orgConversions,
		"recruiter_stats":   recruiterStats,
		"hm_stats":          hmStats,
	})
}
