package db

import (
	"context"
)

const upsertSmtpSettings = `-- name: UpsertSmtpSettings :one
INSERT INTO smtp_settings (organization_id, host, port, username, password, from_email, from_name, tls_enabled)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (organization_id) DO UPDATE SET
    host = EXCLUDED.host,
    port = EXCLUDED.port,
    username = EXCLUDED.username,
    password = EXCLUDED.password,
    from_email = EXCLUDED.from_email,
    from_name = EXCLUDED.from_name,
    tls_enabled = EXCLUDED.tls_enabled,
    updated_at = now()
RETURNING id, organization_id, host, port, username, password, from_email, from_name, tls_enabled, created_at, updated_at
`

type UpsertSmtpSettingsParams struct {
	OrganizationID string `json:"organization_id"`
	Host           string `json:"host"`
	Port           int32  `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	FromEmail      string `json:"from_email"`
	FromName       string `json:"from_name"`
	TlsEnabled     bool   `json:"tls_enabled"`
}

func (q *Queries) UpsertSmtpSettings(ctx context.Context, arg UpsertSmtpSettingsParams) (SmtpSetting, error) {
	row := q.db.QueryRow(ctx, upsertSmtpSettings,
		arg.OrganizationID, arg.Host, arg.Port, arg.Username, arg.Password, arg.FromEmail, arg.FromName, arg.TlsEnabled,
	)
	var i SmtpSetting
	err := row.Scan(
		&i.ID, &i.OrganizationID, &i.Host, &i.Port, &i.Username, &i.Password,
		&i.FromEmail, &i.FromName, &i.TlsEnabled, &i.CreatedAt, &i.UpdatedAt,
	)
	return i, err
}

const getSmtpSettings = `-- name: GetSmtpSettings :one
SELECT id, organization_id, host, port, username, password, from_email, from_name, tls_enabled, created_at, updated_at
FROM smtp_settings WHERE organization_id = $1
`

func (q *Queries) GetSmtpSettings(ctx context.Context, organizationID string) (SmtpSetting, error) {
	row := q.db.QueryRow(ctx, getSmtpSettings, organizationID)
	var i SmtpSetting
	err := row.Scan(
		&i.ID, &i.OrganizationID, &i.Host, &i.Port, &i.Username, &i.Password,
		&i.FromEmail, &i.FromName, &i.TlsEnabled, &i.CreatedAt, &i.UpdatedAt,
	)
	return i, err
}

const deleteSmtpSettings = `-- name: DeleteSmtpSettings :exec
DELETE FROM smtp_settings WHERE organization_id = $1
`

func (q *Queries) DeleteSmtpSettings(ctx context.Context, organizationID string) error {
	_, err := q.db.Exec(ctx, deleteSmtpSettings, organizationID)
	return err
}
