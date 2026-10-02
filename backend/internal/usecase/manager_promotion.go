package usecase

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domainpromotion "github.com/boms/backend/internal/domain/promotion"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	"github.com/boms/backend/internal/shared/utils"
)

// ManagerPromotionUsecase is a manager emailing a promotion to the customers
// who agreed to receive promotions.
type ManagerPromotionUsecase struct {
	promotions port.PromotionRepository
	tx         port.TxManager
	events     port.EventOutbox
	audit      *auditlogger.Service
	log        *zap.Logger
}

func NewManagerPromotionUsecase(
	promotions port.PromotionRepository,
	tx port.TxManager,
	events port.EventOutbox,
	audit *auditlogger.Service,
	log *zap.Logger,
) *ManagerPromotionUsecase {
	return &ManagerPromotionUsecase{promotions: promotions, tx: tx, events: events, audit: audit, log: log}
}

// List returns a page of the promotions sent, latest first.
func (u *ManagerPromotionUsecase) List(ctx context.Context, page, pageSize int32) ([]dto.PromotionResponse, int64, int32, int32, error) {
	page, pageSize = normalizeCatalogListPage(page, pageSize)
	rows, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]port.ManagerPromotion, error) {
			return u.promotions.List(ctx, pageSize, utils.PageOffset(page, pageSize))
		},
		u.promotions.Count,
	)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]dto.PromotionResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toPromotionResponse(row))
	}
	return out, total, page, pageSize, nil
}

// Audience returns how many customers a promotion sent now would go to.
func (u *ManagerPromotionUsecase) Audience(ctx context.Context) (*dto.PromotionAudienceResponse, error) {
	n, err := u.promotions.CountRecipients(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.PromotionAudienceResponse{Recipients: n}, nil
}

// Send records the promotion and, in the same transaction, asks the worker to
// email it to every customer who agreed to promotions when it gets to it.
func (u *ManagerPromotionUsecase) Send(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	req dto.SendPromotionRequest,
) (*dto.PromotionResponse, error) {
	subject, err := domainpromotion.NewSubject(req.Subject)
	if err != nil {
		return nil, err
	}
	body, err := domainpromotion.NewBody(req.Body)
	if err != nil {
		return nil, err
	}
	var created *port.ManagerPromotion
	err = u.tx.WithTx(ctx, func(txCtx context.Context) (err error) {
		created, err = u.promotions.Create(txCtx, port.CreatePromotionParams{Subject: subject, Body: body, CreatedBy: actorID})
		if err != nil {
			return err
		}
		return u.events.Add(txCtx, domainpromotion.CreatedEvent(created.ID))
	})
	if err != nil {
		return nil, err
	}
	recordAudit(u.log, u.audit, ctx, domainpromotion.AuditActionManagerSent, actorID, actorRole, &created.ID, "promotion",
		nil, map[string]string{"subject": subject})
	resp := toPromotionResponse(*created)
	return &resp, nil
}

func toPromotionResponse(p port.ManagerPromotion) dto.PromotionResponse {
	return dto.PromotionResponse{
		ID:             p.ID.String(),
		Subject:        p.Subject,
		Status:         string(p.Status),
		RecipientCount: p.RecipientCount,
		CreatedAt:      p.CreatedAt,
		SenderName:     p.SenderName,
	}
}
