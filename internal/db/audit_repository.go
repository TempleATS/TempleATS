package db

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InsertAuditLog writes an audit log entry. Intended to be called in a goroutine
// (fire-and-forget) so it doesn't slow down request handling.
func InsertAuditLog(ctx context.Context, pool *pgxpool.Pool, orgID, userID, action, entityType, entityID string, details interface{}) {
	var detailsJSON []byte
	if details != nil {
		detailsJSON, _ = json.Marshal(details)
	}

	var uid *string
	if userID != "" {
		uid = &userID
	}

	_, err := pool.Exec(ctx,
		`INSERT INTO audit_logs (user_id, organization_id, action, entity_type, entity_id, details)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		uid, orgID, action, entityType, entityID, detailsJSON,
	)
	if err != nil {
		log.Printf("WARN: failed to write audit log: %v", err)
	}
}
