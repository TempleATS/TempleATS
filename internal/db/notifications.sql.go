package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createNotification = `-- name: CreateNotification :one
INSERT INTO notifications (organization_id, type, recipient_email, recipient_name, subject, body, status, error_message, application_id, note_id, triggered_by_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, organization_id, type, recipient_email, recipient_name, subject, body, status, error_message, application_id, note_id, triggered_by_id, created_at
`

type CreateNotificationParams struct {
	OrganizationID string      `json:"organization_id"`
	Type           string      `json:"type"`
	RecipientEmail string      `json:"recipient_email"`
	RecipientName  string      `json:"recipient_name"`
	Subject        string      `json:"subject"`
	Body           string      `json:"body"`
	Status         string      `json:"status"`
	ErrorMessage   pgtype.Text `json:"error_message"`
	ApplicationID  pgtype.Text `json:"application_id"`
	NoteID         pgtype.Text `json:"note_id"`
	TriggeredByID  pgtype.Text `json:"triggered_by_id"`
}

func (q *Queries) CreateNotification(ctx context.Context, arg CreateNotificationParams) (Notification, error) {
	row := q.db.QueryRow(ctx, createNotification,
		arg.OrganizationID, arg.Type, arg.RecipientEmail, arg.RecipientName,
		arg.Subject, arg.Body, arg.Status, arg.ErrorMessage,
		arg.ApplicationID, arg.NoteID, arg.TriggeredByID,
	)
	var i Notification
	err := row.Scan(
		&i.ID, &i.OrganizationID, &i.Type, &i.RecipientEmail, &i.RecipientName,
		&i.Subject, &i.Body, &i.Status, &i.ErrorMessage,
		&i.ApplicationID, &i.NoteID, &i.TriggeredByID, &i.CreatedAt,
	)
	return i, err
}

const listNotificationsByApplication = `-- name: ListNotificationsByApplication :many
SELECT id, organization_id, type, recipient_email, recipient_name, subject, body, status, error_message, application_id, note_id, triggered_by_id, created_at
FROM notifications WHERE application_id = $1 ORDER BY created_at DESC
`

func (q *Queries) ListNotificationsByApplication(ctx context.Context, applicationID string) ([]Notification, error) {
	rows, err := q.db.Query(ctx, listNotificationsByApplication, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Notification{}
	for rows.Next() {
		var i Notification
		if err := rows.Scan(
			&i.ID, &i.OrganizationID, &i.Type, &i.RecipientEmail, &i.RecipientName,
			&i.Subject, &i.Body, &i.Status, &i.ErrorMessage,
			&i.ApplicationID, &i.NoteID, &i.TriggeredByID, &i.CreatedAt,
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
