package usecase

import (
	"context"

	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/service/auditlogger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func recordAudit(
	log *zap.Logger,
	audit *auditlogger.Service,
	ctx context.Context,
	action domainuser.AuditAction,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	targetID *uuid.UUID,
	targetType string,
	before, after any,
) {
	if audit == nil {
		return
	}
	if err := audit.Log(ctx, action, actorID, actorRole, targetID, targetType, before, after); err != nil {
		if log != nil {
			log.Error("audit_log_failed",
				zap.String("action", string(action)),
				zap.String("target_type", targetType),
				zap.Error(err),
			)
		}
	}
}
