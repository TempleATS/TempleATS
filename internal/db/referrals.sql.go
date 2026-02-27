package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createReferral = `-- name: CreateReferral :one
INSERT INTO referrals (referrer_id, organization_id, job_id, source, candidate_name, candidate_id, application_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, referrer_id, organization_id, job_id, token, source, candidate_name, candidate_id, application_id, created_at
`

type CreateReferralParams struct {
	ReferrerID     string      `json:"referrer_id"`
	OrganizationID string      `json:"organization_id"`
	JobID          string      `json:"job_id"`
	Source         string      `json:"source"`
	CandidateName  pgtype.Text `json:"candidate_name"`
	CandidateID    pgtype.Text `json:"candidate_id"`
	ApplicationID  pgtype.Text `json:"application_id"`
}

func (q *Queries) CreateReferral(ctx context.Context, arg CreateReferralParams) (Referral, error) {
	row := q.db.QueryRow(ctx, createReferral,
		arg.ReferrerID,
		arg.OrganizationID,
		arg.JobID,
		arg.Source,
		arg.CandidateName,
		arg.CandidateID,
		arg.ApplicationID,
	)
	var i Referral
	err := row.Scan(
		&i.ID,
		&i.ReferrerID,
		&i.OrganizationID,
		&i.JobID,
		&i.Token,
		&i.Source,
		&i.CandidateName,
		&i.CandidateID,
		&i.ApplicationID,
		&i.CreatedAt,
	)
	return i, err
}

const getReferralByToken = `-- name: GetReferralByToken :one
SELECT r.id, r.referrer_id, r.organization_id, r.job_id, r.token, r.source,
       r.candidate_name, r.candidate_id, r.application_id, r.created_at,
       u.name AS referrer_name, j.title AS job_title
FROM referrals r
JOIN users u ON r.referrer_id = u.id
JOIN jobs j ON r.job_id = j.id
WHERE r.token = $1
`

type GetReferralByTokenRow struct {
	ID             string             `json:"id"`
	ReferrerID     string             `json:"referrer_id"`
	OrganizationID string             `json:"organization_id"`
	JobID          string             `json:"job_id"`
	Token          string             `json:"token"`
	Source         string             `json:"source"`
	CandidateName  pgtype.Text        `json:"candidate_name"`
	CandidateID    pgtype.Text        `json:"candidate_id"`
	ApplicationID  pgtype.Text        `json:"application_id"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	ReferrerName   string             `json:"referrer_name"`
	JobTitle       string             `json:"job_title"`
}

func (q *Queries) GetReferralByToken(ctx context.Context, token string) (GetReferralByTokenRow, error) {
	row := q.db.QueryRow(ctx, getReferralByToken, token)
	var i GetReferralByTokenRow
	err := row.Scan(
		&i.ID,
		&i.ReferrerID,
		&i.OrganizationID,
		&i.JobID,
		&i.Token,
		&i.Source,
		&i.CandidateName,
		&i.CandidateID,
		&i.ApplicationID,
		&i.CreatedAt,
		&i.ReferrerName,
		&i.JobTitle,
	)
	return i, err
}

const updateReferralApplication = `-- name: UpdateReferralApplication :exec
UPDATE referrals SET candidate_name = $2, candidate_id = $3, application_id = $4
WHERE id = $1
`

type UpdateReferralApplicationParams struct {
	ID            string      `json:"id"`
	CandidateName pgtype.Text `json:"candidate_name"`
	CandidateID   pgtype.Text `json:"candidate_id"`
	ApplicationID pgtype.Text `json:"application_id"`
}

func (q *Queries) UpdateReferralApplication(ctx context.Context, arg UpdateReferralApplicationParams) error {
	_, err := q.db.Exec(ctx, updateReferralApplication, arg.ID, arg.CandidateName, arg.CandidateID, arg.ApplicationID)
	return err
}

const listReferralsByOrg = `-- name: ListReferralsByOrg :many
SELECT r.id, r.source, r.created_at, r.token,
       u.name AS referrer_name,
       r.candidate_name,
       j.title AS job_title, j.id AS job_id,
       COALESCE(a.stage, '') AS application_stage,
       r.application_id
FROM referrals r
JOIN users u ON r.referrer_id = u.id
JOIN jobs j ON r.job_id = j.id
LEFT JOIN applications a ON r.application_id = a.id
WHERE r.organization_id = $1
ORDER BY r.created_at DESC
`

type ListReferralsByOrgRow struct {
	ID               string             `json:"id"`
	Source           string             `json:"source"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Token            string             `json:"token"`
	ReferrerName     string             `json:"referrer_name"`
	CandidateName    pgtype.Text        `json:"candidate_name"`
	JobTitle         string             `json:"job_title"`
	JobID            string             `json:"job_id"`
	ApplicationStage string             `json:"application_stage"`
	ApplicationID    pgtype.Text        `json:"application_id"`
}

func (q *Queries) ListReferralsByOrg(ctx context.Context, organizationID string) ([]ListReferralsByOrgRow, error) {
	rows, err := q.db.Query(ctx, listReferralsByOrg, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ListReferralsByOrgRow, 0)
	for rows.Next() {
		var i ListReferralsByOrgRow
		if err := rows.Scan(
			&i.ID,
			&i.Source,
			&i.CreatedAt,
			&i.Token,
			&i.ReferrerName,
			&i.CandidateName,
			&i.JobTitle,
			&i.JobID,
			&i.ApplicationStage,
			&i.ApplicationID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listReferralsByUser = `-- name: ListReferralsByUser :many
SELECT r.id, r.source, r.created_at, r.token,
       u.name AS referrer_name,
       r.candidate_name,
       j.title AS job_title, j.id AS job_id,
       COALESCE(a.stage, '') AS application_stage,
       r.application_id
FROM referrals r
JOIN users u ON r.referrer_id = u.id
JOIN jobs j ON r.job_id = j.id
LEFT JOIN applications a ON r.application_id = a.id
WHERE r.referrer_id = $1 AND r.organization_id = $2
ORDER BY r.created_at DESC
`

type ListReferralsByUserParams struct {
	ReferrerID     string `json:"referrer_id"`
	OrganizationID string `json:"organization_id"`
}

func (q *Queries) ListReferralsByUser(ctx context.Context, arg ListReferralsByUserParams) ([]ListReferralsByOrgRow, error) {
	rows, err := q.db.Query(ctx, listReferralsByUser, arg.ReferrerID, arg.OrganizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ListReferralsByOrgRow, 0)
	for rows.Next() {
		var i ListReferralsByOrgRow
		if err := rows.Scan(
			&i.ID,
			&i.Source,
			&i.CreatedAt,
			&i.Token,
			&i.ReferrerName,
			&i.CandidateName,
			&i.JobTitle,
			&i.JobID,
			&i.ApplicationStage,
			&i.ApplicationID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
