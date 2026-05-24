package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	"github.com/boms/backend/internal/port"
	"github.com/google/uuid"
)

type AuditLogRepository struct {
	queries *sqlcgen.Queries
}

func NewAuditLogRepository(pool *Pool) *AuditLogRepository {
	return &AuditLogRepository{queries: pool.Queries()}
}

func (r *AuditLogRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *AuditLogRepository) Create(ctx context.Context, params port.CreateAuditLogParams) error {
	targetIDNull := uuid.NullUUID{}
	if params.TargetID != nil {
		targetIDNull = uuid.NullUUID{UUID: *params.TargetID, Valid: true}
	}
	var ip string
	if params.IP != nil {
		ip = *params.IP
	}
	ua := sql.NullString{}
	if params.UserAgent != nil {
		ua = sql.NullString{String: *params.UserAgent, Valid: true}
	}
	err := r.q(ctx).CreateAuditLog(ctx, sqlcgen.CreateAuditLogParams{
		ActorID:     params.ActorID,
		ActorRole:   sqlcgen.UserRole(params.ActorRole),
		Action:      string(params.Action),
		TargetID:    targetIDNull,
		TargetType:  params.TargetType,
		BeforeJsonb: json.RawMessage(params.BeforeJSON),
		AfterJsonb:  json.RawMessage(params.AfterJSON),
		Column8:     ip,
		UserAgent:   ua,
	})
	return mapUserQueryError(err, "create audit log")
}

var _ port.AuditLogRepository = (*AuditLogRepository)(nil)
