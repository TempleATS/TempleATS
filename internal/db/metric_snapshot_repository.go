package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MetricSnapshot represents a saved point-in-time metrics snapshot for a requisition.
type MetricSnapshot struct {
	ID               string    `json:"id"`
	RequisitionID    string    `json:"requisition_id"`
	OrganizationID   string    `json:"organization_id"`
	CreatedByID      string    `json:"created_by_id"`
	Label            *string   `json:"label"`
	FunnelApplied    int       `json:"funnel_applied"`
	FunnelHRScreen   int       `json:"funnel_hr_screen"`
	FunnelHMReview   int       `json:"funnel_hm_review"`
	FunnelFirstIV    int       `json:"funnel_first_interview"`
	FunnelFinalIV    int       `json:"funnel_final_interview"`
	FunnelOffer      int       `json:"funnel_offer"`
	FunnelHired      int       `json:"funnel_hired"`
	FunnelRejected   int       `json:"funnel_rejected"`
	TPApplied        int       `json:"tp_applied"`
	TPFirstIV        int       `json:"tp_first_interview"`
	TPFinalIV        int       `json:"tp_final_interview"`
	TPOffer          int       `json:"tp_offer"`
	TPHired          int       `json:"tp_hired"`
	RatioFirstFinal  *float64  `json:"ratio_first_to_final"`
	RatioFinalOffer  *float64  `json:"ratio_final_to_offer"`
	RatioOfferHired  *float64  `json:"ratio_offer_to_hired"`
	TotalApps        int       `json:"total_applications"`
	TargetHires      int       `json:"target_hires"`
	CreatedAt        time.Time `json:"created_at"`
}

// ThroughputRow is a stage + count pair from the throughput query.
type ThroughputRow struct {
	Stage string
	Count int
}

// PersonStats holds aggregated metrics for a recruiter or hiring manager.
type PersonStats struct {
	UserID           string   `json:"user_id"`
	UserName         string   `json:"user_name"`
	TotalReqs        int      `json:"total_reqs"`
	OpenReqs         int      `json:"open_reqs"`
	TotalApps        int      `json:"total_applications"`
	TotalHired       int      `json:"total_hired"`
	TotalRejected    int      `json:"total_rejected"`
	TPFirstIV        int      `json:"tp_first_interview"`
	TPFinalIV        int      `json:"tp_final_interview"`
	TPOffer          int      `json:"tp_offer"`
	TPHired          int      `json:"tp_hired"`
	RatioFirstFinal  *float64 `json:"ratio_first_to_final"`
	RatioFinalOffer  *float64 `json:"ratio_final_to_offer"`
	RatioOfferHired  *float64 `json:"ratio_offer_to_hired"`
}

// MetricSnapshotRepository handles metric snapshot queries.
type MetricSnapshotRepository struct {
	pool *pgxpool.Pool
}

// NewMetricSnapshotRepository creates a new repository.
func NewMetricSnapshotRepository(pool *pgxpool.Pool) *MetricSnapshotRepository {
	return &MetricSnapshotRepository{pool: pool}
}

// ThroughputByRequisition counts distinct candidates who ever entered each stage.
func (r *MetricSnapshotRepository) ThroughputByRequisition(ctx context.Context, reqID string) ([]ThroughputRow, error) {
	query := `
		SELECT st.to_stage AS stage, COUNT(DISTINCT a.candidate_id)::int AS cnt
		FROM stage_transitions st
		JOIN applications a ON st.application_id = a.id
		JOIN jobs j ON a.job_id = j.id
		WHERE j.requisition_id = $1
		GROUP BY st.to_stage
		UNION ALL
		SELECT 'applied' AS stage, COUNT(DISTINCT a.candidate_id)::int AS cnt
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE j.requisition_id = $1
	`
	rows, err := r.pool.Query(ctx, query, reqID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ThroughputRow
	for rows.Next() {
		var tr ThroughputRow
		if err := rows.Scan(&tr.Stage, &tr.Count); err != nil {
			return nil, err
		}
		result = append(result, tr)
	}
	return result, rows.Err()
}

// InsertSnapshot saves a new metric snapshot.
func (r *MetricSnapshotRepository) InsertSnapshot(ctx context.Context, s *MetricSnapshot) error {
	query := `
		INSERT INTO metric_snapshots (
			requisition_id, organization_id, created_by_id, label,
			funnel_applied, funnel_hr_screen, funnel_hm_review,
			funnel_first_interview, funnel_final_interview, funnel_offer,
			funnel_hired, funnel_rejected,
			tp_applied, tp_first_interview, tp_final_interview, tp_offer, tp_hired,
			ratio_first_to_final, ratio_final_to_offer, ratio_offer_to_hired,
			total_applications, target_hires
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17,
			$18, $19, $20,
			$21, $22
		) RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		s.RequisitionID, s.OrganizationID, s.CreatedByID, s.Label,
		s.FunnelApplied, s.FunnelHRScreen, s.FunnelHMReview,
		s.FunnelFirstIV, s.FunnelFinalIV, s.FunnelOffer,
		s.FunnelHired, s.FunnelRejected,
		s.TPApplied, s.TPFirstIV, s.TPFinalIV, s.TPOffer, s.TPHired,
		s.RatioFirstFinal, s.RatioFinalOffer, s.RatioOfferHired,
		s.TotalApps, s.TargetHires,
	).Scan(&s.ID, &s.CreatedAt)
}

// ListSnapshotsByReq returns all snapshots for a requisition, newest first.
func (r *MetricSnapshotRepository) ListSnapshotsByReq(ctx context.Context, reqID string) ([]MetricSnapshot, error) {
	query := `
		SELECT id, requisition_id, organization_id, created_by_id, label,
			funnel_applied, funnel_hr_screen, funnel_hm_review,
			funnel_first_interview, funnel_final_interview, funnel_offer,
			funnel_hired, funnel_rejected,
			tp_applied, tp_first_interview, tp_final_interview, tp_offer, tp_hired,
			ratio_first_to_final, ratio_final_to_offer, ratio_offer_to_hired,
			total_applications, target_hires, created_at
		FROM metric_snapshots
		WHERE requisition_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, reqID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MetricSnapshot
	for rows.Next() {
		var s MetricSnapshot
		if err := rows.Scan(
			&s.ID, &s.RequisitionID, &s.OrganizationID, &s.CreatedByID, &s.Label,
			&s.FunnelApplied, &s.FunnelHRScreen, &s.FunnelHMReview,
			&s.FunnelFirstIV, &s.FunnelFinalIV, &s.FunnelOffer,
			&s.FunnelHired, &s.FunnelRejected,
			&s.TPApplied, &s.TPFirstIV, &s.TPFinalIV, &s.TPOffer, &s.TPHired,
			&s.RatioFirstFinal, &s.RatioFinalOffer, &s.RatioOfferHired,
			&s.TotalApps, &s.TargetHires, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// DeleteSnapshot deletes a snapshot by ID.
func (r *MetricSnapshotRepository) DeleteSnapshot(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM metric_snapshots WHERE id = $1`, id)
	return err
}

// RecruiterStats returns per-recruiter aggregated metrics for an org.
func (r *MetricSnapshotRepository) RecruiterStats(ctx context.Context, orgID string) ([]PersonStats, error) {
	return r.personStats(ctx, orgID, "recruiter_id")
}

// HiringManagerStats returns per-hiring-manager aggregated metrics for an org.
func (r *MetricSnapshotRepository) HiringManagerStats(ctx context.Context, orgID string) ([]PersonStats, error) {
	return r.personStats(ctx, orgID, "hiring_manager_id")
}

func (r *MetricSnapshotRepository) personStats(ctx context.Context, orgID, roleColumn string) ([]PersonStats, error) {
	// Build query - roleColumn is always a hardcoded string ("recruiter_id" or "hiring_manager_id"), safe to interpolate
	query := `
		WITH req_apps AS (
			SELECT r.id AS req_id, r.` + roleColumn + ` AS person_id, r.status AS req_status,
				a.id AS app_id, a.stage, a.candidate_id
			FROM requisitions r
			JOIN jobs j ON j.requisition_id = r.id
			JOIN applications a ON a.job_id = j.id
			WHERE r.organization_id = $1 AND r.` + roleColumn + ` IS NOT NULL
		),
		req_transitions AS (
			SELECT r.` + roleColumn + ` AS person_id, st.to_stage, a.candidate_id
			FROM requisitions r
			JOIN jobs j ON j.requisition_id = r.id
			JOIN applications a ON a.job_id = j.id
			JOIN stage_transitions st ON st.application_id = a.id
			WHERE r.organization_id = $1 AND r.` + roleColumn + ` IS NOT NULL
		)
		SELECT
			u.id,
			u.name,
			COUNT(DISTINCT ra.req_id)::int AS total_reqs,
			COUNT(DISTINCT CASE WHEN ra.req_status = 'open' THEN ra.req_id END)::int AS open_reqs,
			COUNT(DISTINCT ra.app_id)::int AS total_apps,
			COUNT(DISTINCT CASE WHEN ra.stage = 'hired' THEN ra.app_id END)::int AS total_hired,
			COUNT(DISTINCT CASE WHEN ra.stage = 'rejected' THEN ra.app_id END)::int AS total_rejected,
			COALESCE((SELECT COUNT(DISTINCT candidate_id)::int FROM req_transitions WHERE person_id = u.id AND to_stage = 'first_interview'), 0),
			COALESCE((SELECT COUNT(DISTINCT candidate_id)::int FROM req_transitions WHERE person_id = u.id AND to_stage = 'final_interview'), 0),
			COALESCE((SELECT COUNT(DISTINCT candidate_id)::int FROM req_transitions WHERE person_id = u.id AND to_stage = 'offer'), 0),
			COALESCE((SELECT COUNT(DISTINCT candidate_id)::int FROM req_transitions WHERE person_id = u.id AND to_stage = 'hired'), 0)
		FROM users u
		JOIN req_apps ra ON ra.person_id = u.id
		GROUP BY u.id, u.name
		ORDER BY total_apps DESC
	`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PersonStats
	for rows.Next() {
		var ps PersonStats
		if err := rows.Scan(
			&ps.UserID, &ps.UserName,
			&ps.TotalReqs, &ps.OpenReqs,
			&ps.TotalApps, &ps.TotalHired, &ps.TotalRejected,
			&ps.TPFirstIV, &ps.TPFinalIV, &ps.TPOffer, &ps.TPHired,
		); err != nil {
			return nil, err
		}
		// Compute ratios
		if ps.TPFirstIV > 0 {
			v := float64(ps.TPFinalIV) / float64(ps.TPFirstIV)
			ps.RatioFirstFinal = &v
		}
		if ps.TPFinalIV > 0 {
			v := float64(ps.TPOffer) / float64(ps.TPFinalIV)
			ps.RatioFinalOffer = &v
		}
		if ps.TPOffer > 0 {
			v := float64(ps.TPHired) / float64(ps.TPOffer)
			ps.RatioOfferHired = &v
		}
		result = append(result, ps)
	}
	return result, rows.Err()
}

// OrgThroughput returns org-wide throughput counts across all requisitions.
func (r *MetricSnapshotRepository) OrgThroughput(ctx context.Context, orgID string) (map[string]int, error) {
	query := `
		SELECT st.to_stage AS stage, COUNT(DISTINCT a.candidate_id)::int AS cnt
		FROM stage_transitions st
		JOIN applications a ON st.application_id = a.id
		JOIN jobs j ON a.job_id = j.id
		JOIN requisitions r ON j.requisition_id = r.id
		WHERE r.organization_id = $1
		GROUP BY st.to_stage
		UNION ALL
		SELECT 'applied' AS stage, COUNT(DISTINCT a.candidate_id)::int AS cnt
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		JOIN requisitions r ON j.requisition_id = r.id
		WHERE r.organization_id = $1
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]int{}
	for rows.Next() {
		var stage string
		var cnt int
		if err := rows.Scan(&stage, &cnt); err != nil {
			return nil, err
		}
		result[stage] = cnt
	}
	return result, rows.Err()
}

// Ensure pgx.Rows is closed properly — satisfy any interface if needed.
var _ pgx.Rows = (pgx.Rows)(nil)
