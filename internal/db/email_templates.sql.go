package db

import (
	"context"
)

const upsertEmailTemplate = `-- name: UpsertEmailTemplate :one
INSERT INTO email_templates (organization_id, stage, subject, body, enabled)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (organization_id, stage) DO UPDATE SET
    subject = EXCLUDED.subject,
    body = EXCLUDED.body,
    enabled = EXCLUDED.enabled,
    updated_at = now()
RETURNING id, organization_id, stage, subject, body, enabled, created_at, updated_at
`

type UpsertEmailTemplateParams struct {
	OrganizationID string `json:"organization_id"`
	Stage          string `json:"stage"`
	Subject        string `json:"subject"`
	Body           string `json:"body"`
	Enabled        bool   `json:"enabled"`
}

func (q *Queries) UpsertEmailTemplate(ctx context.Context, arg UpsertEmailTemplateParams) (EmailTemplate, error) {
	row := q.db.QueryRow(ctx, upsertEmailTemplate,
		arg.OrganizationID, arg.Stage, arg.Subject, arg.Body, arg.Enabled,
	)
	var i EmailTemplate
	err := row.Scan(
		&i.ID, &i.OrganizationID, &i.Stage, &i.Subject, &i.Body, &i.Enabled, &i.CreatedAt, &i.UpdatedAt,
	)
	return i, err
}

const listEmailTemplatesByOrg = `-- name: ListEmailTemplatesByOrg :many
SELECT id, organization_id, stage, subject, body, enabled, created_at, updated_at
FROM email_templates WHERE organization_id = $1 ORDER BY stage
`

func (q *Queries) ListEmailTemplatesByOrg(ctx context.Context, organizationID string) ([]EmailTemplate, error) {
	rows, err := q.db.Query(ctx, listEmailTemplatesByOrg, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []EmailTemplate{}
	for rows.Next() {
		var i EmailTemplate
		if err := rows.Scan(
			&i.ID, &i.OrganizationID, &i.Stage, &i.Subject, &i.Body, &i.Enabled, &i.CreatedAt, &i.UpdatedAt,
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

const getEmailTemplateByStage = `-- name: GetEmailTemplateByStage :one
SELECT id, organization_id, stage, subject, body, enabled, created_at, updated_at
FROM email_templates WHERE organization_id = $1 AND stage = $2
`

type GetEmailTemplateByStageParams struct {
	OrganizationID string `json:"organization_id"`
	Stage          string `json:"stage"`
}

func (q *Queries) GetEmailTemplateByStage(ctx context.Context, arg GetEmailTemplateByStageParams) (EmailTemplate, error) {
	row := q.db.QueryRow(ctx, getEmailTemplateByStage, arg.OrganizationID, arg.Stage)
	var i EmailTemplate
	err := row.Scan(
		&i.ID, &i.OrganizationID, &i.Stage, &i.Subject, &i.Body, &i.Enabled, &i.CreatedAt, &i.UpdatedAt,
	)
	return i, err
}
