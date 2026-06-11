package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainuser "github.com/boms/backend/internal/domain/user"
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
	return mapRepoError(err, "create audit log")
}

func (r *AuditLogRepository) CountByTargetID(ctx context.Context, targetID uuid.UUID) (int64, error) {
	total, err := r.q(ctx).CountAuditLogsByTargetID(ctx, uuid.NullUUID{UUID: targetID, Valid: true})
	if err != nil {
		return 0, mapRepoError(err, "count audit logs by target")
	}
	return total, nil
}

func (r *AuditLogRepository) ListByTargetID(
	ctx context.Context,
	params port.ListAuditLogsByTargetParams,
) ([]port.AuditLogEntry, error) {
	rows, err := r.q(ctx).ListAuditLogsByTargetID(ctx, sqlcgen.ListAuditLogsByTargetIDParams{
		TargetID: uuid.NullUUID{UUID: params.TargetID, Valid: true},
		Limit:    params.Limit,
		Offset:   params.Offset,
	})
	if err != nil {
		return nil, mapRepoError(err, "list audit logs by target")
	}

	out := make([]port.AuditLogEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.AuditLogEntry{
			ID:         row.ID,
			ActorID:    row.ActorID,
			ActorRole:  domainuser.Role(row.ActorRole),
			ActorEmail: row.ActorEmail,
			Action:     domainuser.AuditAction(row.Action),
			BeforeJSON: row.BeforeJsonb,
			AfterJSON:  row.AfterJsonb,
			CreatedAt:  row.CreatedAt,
		})
	}
	return out, nil
}

var _ port.AuditLogRepository = (*AuditLogRepository)(nil)
