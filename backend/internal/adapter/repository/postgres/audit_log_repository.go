package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
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
	var ip string
	if params.IP != nil {
		ip = *params.IP
	}
	err := r.q(ctx).CreateAuditLog(ctx, sqlcgen.CreateAuditLogParams{
		ActorID:     params.ActorID,
		ActorRole:   sqlcgen.UserRole(params.ActorRole),
		Action:      string(params.Action),
		TargetID:    params.TargetID,
		TargetType:  params.TargetType,
		BeforeJsonb: json.RawMessage(params.BeforeJSON),
		AfterJsonb:  json.RawMessage(params.AfterJSON),
		Column8:     ip,
		UserAgent:   params.UserAgent,
	})
	if err != nil {
		return mapRepoError(err, "create audit log")
	}
	return nil
}

func (r *AuditLogRepository) CountByTargetID(ctx context.Context, targetID uuid.UUID) (int64, error) {
	total, err := r.q(ctx).CountAuditLogsByTargetID(ctx, &targetID)
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
		TargetID: &params.TargetID,
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

// ListForSubject implements port.AccountActivityReader.
func (r *AuditLogRepository) ListForSubject(
	ctx context.Context,
	subjectID uuid.UUID,
	before *port.PageCursor,
	limit int32,
) ([]port.AccountActivity, error) {
	params := sqlcgen.ListAuditLogsForSubjectBeforeParams{SubjectID: &subjectID, Limit: limit}
	if before != nil {
		params.BeforeAt, params.BeforeID = &before.At, &before.ID
	}
	rows, err := r.q(ctx).ListAuditLogsForSubjectBefore(ctx, params)
	if err != nil {
		return nil, mapRepoError(err, "list audit logs for subject")
	}
	out := make([]port.AccountActivity, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.AccountActivity{
			ID:         row.ID,
			ActorID:    row.ActorID,
			ActorRole:  domainuser.Role(row.ActorRole),
			Action:     domainuser.AuditAction(row.Action),
			BeforeJSON: row.BeforeJsonb,
			AfterJSON:  row.AfterJsonb,
			IP:         row.Ip,
			UserAgent:  row.UserAgent,
			At:         row.CreatedAt,
		})
	}
	return out, nil
}

// ScrubSubject implements port.AuditScrubber.
func (r *AuditLogRepository) ScrubSubject(ctx context.Context, subjectID uuid.UUID) error {
	if err := r.q(ctx).ScrubAuditLogsForSubject(ctx, &subjectID); err != nil {
		return mapRepoError(err, "scrub audit logs for subject")
	}
	return nil
}

var (
	_ port.AccountActivityReader = (*AuditLogRepository)(nil)
	_ port.AuditScrubber         = (*AuditLogRepository)(nil)
)
