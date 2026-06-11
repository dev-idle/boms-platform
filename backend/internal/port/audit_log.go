package port

import (
	"context"
	"time"

	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/google/uuid"
)

type CreateAuditLogParams struct {
	ActorID    uuid.UUID
	ActorRole  domainuser.Role
	Action     domainuser.AuditAction
	TargetID   *uuid.UUID
	TargetType string
	BeforeJSON []byte
	AfterJSON  []byte
	IP         *string
	UserAgent  *string
}

type AuditLogEntry struct {
	ID         uuid.UUID
	ActorID    uuid.UUID
	ActorRole  domainuser.Role
	ActorEmail string
	Action     domainuser.AuditAction
	BeforeJSON []byte
	AfterJSON  []byte
	CreatedAt  time.Time
}

type ListAuditLogsByTargetParams struct {
	TargetID uuid.UUID
	Limit    int32
	Offset   int32
}

type AuditLogRepository interface {
	Create(ctx context.Context, params CreateAuditLogParams) error
	CountByTargetID(ctx context.Context, targetID uuid.UUID) (int64, error)
	ListByTargetID(ctx context.Context, params ListAuditLogsByTargetParams) ([]AuditLogEntry, error)
}
