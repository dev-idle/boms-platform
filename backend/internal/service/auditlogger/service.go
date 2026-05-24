package auditlogger

import (
	"context"
	"encoding/json"

	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/shared/ctxmeta"
	"github.com/google/uuid"
)

type Service struct {
	repo port.AuditLogRepository
}

func NewService(repo port.AuditLogRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Log(
	ctx context.Context,
	action domainuser.AuditAction,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	targetID *uuid.UUID,
	targetType string,
	before any,
	after any,
) error {
	beforeJSON, err := normalizeJSON(before)
	if err != nil {
		return err
	}
	afterJSON, err := normalizeJSON(after)
	if err != nil {
		return err
	}

	var ipPtr, uaPtr *string
	if ip := ctxmeta.IP(ctx); ip != "" {
		ipVal := ip
		ipPtr = &ipVal
	}
	if ua := ctxmeta.UserAgent(ctx); ua != "" {
		uaVal := ua
		uaPtr = &uaVal
	}

	return s.repo.Create(ctx, port.CreateAuditLogParams{
		ActorID:    actorID,
		ActorRole:  actorRole,
		Action:     action,
		TargetID:   targetID,
		TargetType: targetType,
		BeforeJSON: beforeJSON,
		AfterJSON:  afterJSON,
		IP:         ipPtr,
		UserAgent:  uaPtr,
	})
}

func normalizeJSON(v any) ([]byte, error) {
	if v == nil {
		return []byte("{}"), nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if string(b) == "null" {
		return []byte("{}"), nil
	}
	return b, nil
}
