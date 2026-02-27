package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// JobSearchResult mirrors the jobs table columns needed for list views.
type JobSearchResult struct {
	ID             string             `json:"id"`
	Title          string             `json:"title"`
	Location       pgtype.Text        `json:"location"`
	Department     pgtype.Text        `json:"department"`
	Salary         pgtype.Text        `json:"salary"`
	Status         string             `json:"status"`
	RequisitionID  pgtype.Text        `json:"requisition_id"`
	OrganizationID string             `json:"organization_id"`
	CreatedAt      time.Time          `json:"created_at"`
	Rank           float64            `json:"rank,omitempty"`
}

// ReqSearchResult mirrors the requisitions table columns needed for list views.
type ReqSearchResult struct {
	ID              string             `json:"id"`
	Title           string             `json:"title"`
	JobCode         pgtype.Text        `json:"job_code"`
	Level           pgtype.Text        `json:"level"`
	Department      pgtype.Text        `json:"department"`
	TargetHires     int32              `json:"target_hires"`
	Status          string             `json:"status"`
	HiringManagerID pgtype.Text        `json:"hiring_manager_id"`
	RecruiterID     pgtype.Text        `json:"recruiter_id"`
	OrganizationID  string             `json:"organization_id"`
	OpenedAt        time.Time          `json:"opened_at"`
	ClosedAt        pgtype.Timestamptz `json:"closed_at"`
	Rank            float64            `json:"rank,omitempty"`
}

// SearchJobs searches jobs by full-text search with trigram fallback.
// Returns results ranked by relevance: full-text matches first, then fuzzy.
func SearchJobs(ctx context.Context, pool *pgxpool.Pool, orgID, query string) ([]JobSearchResult, error) {
	rows, err := pool.Query(ctx, `
		WITH fts AS (
			SELECT id, title, location, department, salary, status,
			       requisition_id, organization_id, created_at,
			       ts_rank(search_vector, websearch_to_tsquery('english', $2)) AS rank
			FROM jobs
			WHERE organization_id = $1
			  AND search_vector @@ websearch_to_tsquery('english', $2)
		),
		trgm AS (
			SELECT id, title, location, department, salary, status,
			       requisition_id, organization_id, created_at,
			       similarity(title, $2) AS rank
			FROM jobs
			WHERE organization_id = $1
			  AND similarity(title, $2) > 0.1
			  AND id NOT IN (SELECT id FROM fts)
		)
		SELECT * FROM fts
		UNION ALL
		SELECT * FROM trgm
		ORDER BY rank DESC
	`, orgID, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []JobSearchResult
	for rows.Next() {
		var r JobSearchResult
		err := rows.Scan(
			&r.ID, &r.Title, &r.Location, &r.Department, &r.Salary,
			&r.Status, &r.RequisitionID, &r.OrganizationID, &r.CreatedAt, &r.Rank,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if results == nil {
		results = []JobSearchResult{}
	}
	return results, rows.Err()
}

// SearchJobsByHiringManager searches jobs scoped to a hiring manager.
func SearchJobsByHiringManager(ctx context.Context, pool *pgxpool.Pool, orgID, hmID, query string) ([]JobSearchResult, error) {
	rows, err := pool.Query(ctx, `
		WITH fts AS (
			SELECT j.id, j.title, j.location, j.department, j.salary, j.status,
			       j.requisition_id, j.organization_id, j.created_at,
			       ts_rank(j.search_vector, websearch_to_tsquery('english', $3)) AS rank
			FROM jobs j
			JOIN requisitions r ON j.requisition_id = r.id
			WHERE j.organization_id = $1
			  AND r.hiring_manager_id = $2
			  AND j.search_vector @@ websearch_to_tsquery('english', $3)
		),
		trgm AS (
			SELECT j.id, j.title, j.location, j.department, j.salary, j.status,
			       j.requisition_id, j.organization_id, j.created_at,
			       similarity(j.title, $3) AS rank
			FROM jobs j
			JOIN requisitions r ON j.requisition_id = r.id
			WHERE j.organization_id = $1
			  AND r.hiring_manager_id = $2
			  AND similarity(j.title, $3) > 0.1
			  AND j.id NOT IN (SELECT id FROM fts)
		)
		SELECT * FROM fts
		UNION ALL
		SELECT * FROM trgm
		ORDER BY rank DESC
	`, orgID, hmID, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []JobSearchResult
	for rows.Next() {
		var r JobSearchResult
		err := rows.Scan(
			&r.ID, &r.Title, &r.Location, &r.Department, &r.Salary,
			&r.Status, &r.RequisitionID, &r.OrganizationID, &r.CreatedAt, &r.Rank,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if results == nil {
		results = []JobSearchResult{}
	}
	return results, rows.Err()
}

// SearchRequisitions searches requisitions by full-text search with trigram fallback.
func SearchRequisitions(ctx context.Context, pool *pgxpool.Pool, orgID, query string) ([]ReqSearchResult, error) {
	rows, err := pool.Query(ctx, `
		WITH fts AS (
			SELECT id, title, job_code, level, department, target_hires,
			       status, hiring_manager_id, recruiter_id, organization_id,
			       opened_at, closed_at,
			       ts_rank(search_vector, websearch_to_tsquery('english', $2)) AS rank
			FROM requisitions
			WHERE organization_id = $1
			  AND search_vector @@ websearch_to_tsquery('english', $2)
		),
		trgm AS (
			SELECT id, title, job_code, level, department, target_hires,
			       status, hiring_manager_id, recruiter_id, organization_id,
			       opened_at, closed_at,
			       similarity(title, $2) AS rank
			FROM requisitions
			WHERE organization_id = $1
			  AND similarity(title, $2) > 0.1
			  AND id NOT IN (SELECT id FROM fts)
		)
		SELECT * FROM fts
		UNION ALL
		SELECT * FROM trgm
		ORDER BY rank DESC
	`, orgID, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ReqSearchResult
	for rows.Next() {
		var r ReqSearchResult
		err := rows.Scan(
			&r.ID, &r.Title, &r.JobCode, &r.Level, &r.Department, &r.TargetHires,
			&r.Status, &r.HiringManagerID, &r.RecruiterID, &r.OrganizationID,
			&r.OpenedAt, &r.ClosedAt, &r.Rank,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if results == nil {
		results = []ReqSearchResult{}
	}
	return results, rows.Err()
}

// SearchRequisitionsByHiringManager searches requisitions scoped to a hiring manager.
func SearchRequisitionsByHiringManager(ctx context.Context, pool *pgxpool.Pool, orgID, hmID, query string) ([]ReqSearchResult, error) {
	rows, err := pool.Query(ctx, `
		WITH fts AS (
			SELECT id, title, job_code, level, department, target_hires,
			       status, hiring_manager_id, recruiter_id, organization_id,
			       opened_at, closed_at,
			       ts_rank(search_vector, websearch_to_tsquery('english', $3)) AS rank
			FROM requisitions
			WHERE organization_id = $1
			  AND hiring_manager_id = $2
			  AND search_vector @@ websearch_to_tsquery('english', $3)
		),
		trgm AS (
			SELECT id, title, job_code, level, department, target_hires,
			       status, hiring_manager_id, recruiter_id, organization_id,
			       opened_at, closed_at,
			       similarity(title, $3) AS rank
			FROM requisitions
			WHERE organization_id = $1
			  AND hiring_manager_id = $2
			  AND similarity(title, $3) > 0.1
			  AND id NOT IN (SELECT id FROM fts)
		)
		SELECT * FROM fts
		UNION ALL
		SELECT * FROM trgm
		ORDER BY rank DESC
	`, orgID, hmID, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ReqSearchResult
	for rows.Next() {
		var r ReqSearchResult
		err := rows.Scan(
			&r.ID, &r.Title, &r.JobCode, &r.Level, &r.Department, &r.TargetHires,
			&r.Status, &r.HiringManagerID, &r.RecruiterID, &r.OrganizationID,
			&r.OpenedAt, &r.ClosedAt, &r.Rank,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if results == nil {
		results = []ReqSearchResult{}
	}
	return results, rows.Err()
}
