package port

import (
	"context"

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

type AuditLogRepository interface {
	Create(ctx context.Context, params CreateAuditLogParams) error
}
