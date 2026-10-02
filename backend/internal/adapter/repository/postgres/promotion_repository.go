package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainpromotion "github.com/boms/backend/internal/domain/promotion"
	"github.com/boms/backend/internal/port"
)

// PromotionRepository keeps the promotions managers send and finds the
// customers they go to.
type PromotionRepository struct {
	queries *sqlcgen.Queries
}

func NewPromotionRepository(pool *Pool) *PromotionRepository {
	return &PromotionRepository{queries: pool.Queries()}
}

func (r *PromotionRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// Create implements port.PromotionRepository.
func (r *PromotionRepository) Create(ctx context.Context, params port.CreatePromotionParams) (*port.ManagerPromotion, error) {
	row, err := r.q(ctx).CreatePromotion(ctx, sqlcgen.CreatePromotionParams{
		Subject:   params.Subject,
		Body:      params.Body,
		CreatedBy: params.CreatedBy,
	})
	if err != nil {
		return nil, mapRepoError(err, "create promotion")
	}
	promotion := mapManagerPromotion(sqlcgen.ListPromotionsRow(row))
	return &promotion, nil
}

// List implements port.PromotionRepository.
func (r *PromotionRepository) List(ctx context.Context, limit, offset int32) ([]port.ManagerPromotion, error) {
	rows, err := r.q(ctx).ListPromotions(ctx, sqlcgen.ListPromotionsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, mapRepoError(err, "list promotions")
	}
	out := make([]port.ManagerPromotion, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapManagerPromotion(row))
	}
	return out, nil
}

// Count implements port.PromotionRepository.
func (r *PromotionRepository) Count(ctx context.Context) (int64, error) {
	n, err := r.q(ctx).CountPromotions(ctx)
	if err != nil {
		return 0, mapRepoError(err, "count promotions")
	}
	return n, nil
}

// GetStatus implements port.PromotionRepository.
func (r *PromotionRepository) GetStatus(ctx context.Context, id uuid.UUID) (domainpromotion.Status, error) {
	status, err := r.q(ctx).GetPromotionStatus(ctx, id)
	if err != nil {
		return "", mapRepoError(err, "get promotion status")
	}
	return domainpromotion.Status(status), nil
}

// ListRecipients implements port.PromotionRepository.
func (r *PromotionRepository) ListRecipients(ctx context.Context, afterID *uuid.UUID, limit int32) ([]uuid.UUID, error) {
	ids, err := r.q(ctx).ListPromotionRecipients(ctx, sqlcgen.ListPromotionRecipientsParams{AfterID: afterID, Limit: limit})
	if err != nil {
		return nil, mapRepoError(err, "list promotion recipients")
	}
	return ids, nil
}

// CountRecipients implements port.PromotionRepository.
func (r *PromotionRepository) CountRecipients(ctx context.Context) (int64, error) {
	n, err := r.q(ctx).CountPromotionRecipients(ctx)
	if err != nil {
		return 0, mapRepoError(err, "count promotion recipients")
	}
	return n, nil
}

// GetDelivery implements port.PromotionRepository.
func (r *PromotionRepository) GetDelivery(ctx context.Context, promotionID, userID uuid.UUID) (*port.PromotionDelivery, error) {
	row, err := r.q(ctx).GetPromotionDelivery(ctx, sqlcgen.GetPromotionDeliveryParams{PromotionID: promotionID, UserID: userID})
	if err != nil {
		return nil, mapRepoError(err, "get promotion delivery")
	}
	return &port.PromotionDelivery{Subject: row.Subject, Body: row.Body, Email: row.Email, Name: row.DisplayName}, nil
}

// MarkSent implements port.PromotionRepository.
func (r *PromotionRepository) MarkSent(ctx context.Context, id uuid.UUID, recipients int32) (bool, error) {
	n, err := r.q(ctx).MarkPromotionSent(ctx, sqlcgen.MarkPromotionSentParams{ID: id, RecipientCount: &recipients})
	if err != nil {
		return false, mapRepoError(err, "mark promotion sent")
	}
	return n > 0, nil
}

func mapManagerPromotion(row sqlcgen.ListPromotionsRow) port.ManagerPromotion {
	return port.ManagerPromotion{
		ID:             row.ID,
		Subject:        row.Subject,
		Status:         domainpromotion.Status(row.Status),
		RecipientCount: row.RecipientCount,
		CreatedAt:      row.CreatedAt,
		SenderName:     row.SenderName,
	}
}

var _ port.PromotionRepository = (*PromotionRepository)(nil)
