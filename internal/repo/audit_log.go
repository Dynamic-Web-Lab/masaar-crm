package repo

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maidulcu/masaar-crm/internal/domain"
)

type AuditLogRepo struct {
	db *pgxpool.Pool
}

func NewAuditLogRepo(db *pgxpool.Pool) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

// Log inserts an audit entry. Errors are intentionally non-fatal — a failed
// audit write must not block the primary operation.
func (r *AuditLogRepo) Log(ctx context.Context, actorID uuid.UUID, action, entityType string, entityID uuid.UUID, diff any) {
	var diffJSON []byte
	if diff != nil {
		diffJSON, _ = json.Marshal(diff)
	}
	const q = `
		INSERT INTO audit_logs (entity_type, entity_id, action, actor_id, diff)
		VALUES ($1, $2, $3, $4, $5)
	`
	r.db.Exec(ctx, q, entityType, entityID, action, actorID, diffJSON)
}

// AuditLog constants used across handlers.
const (
	AuditCreate         = "create"
	AuditUpdate         = "update"
	AuditDelete         = "delete"
	AuditLogin          = "login"
	AuditLogout         = "logout"
	AuditPasswordChange = "password_change"

	AuditContact = domain.AuditEntityContact
	AuditLead    = domain.AuditEntityLead
	AuditDeal    = domain.AuditEntityDeal
	AuditInvoice = domain.AuditEntityInvoice
	AuditUser     = "user"
	AuditDocument = "document"
)
