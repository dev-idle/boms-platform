package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainpromotion "github.com/boms/backend/internal/domain/promotion"
)

// PromotionRepository keeps the promotions managers send and finds the
// customers they go to: open accounts with a confirmed address whose holder
// agreed to promotions.
type PromotionRepository interface {
	Create(ctx context.Context, params CreatePromotionParams) (*ManagerPromotion, error)
	List(ctx context.Context, limit, offset int32) ([]ManagerPromotion, error)
	Count(ctx context.Context) (int64, error)
	GetStatus(ctx context.Context, id uuid.UUID) (domainpromotion.Status, error)
	// ListRecipients pages the customers promotions go to by id, after afterID.
	ListRecipients(ctx context.Context, afterID *uuid.UUID, limit int32) ([]uuid.UUID, error)
	CountRecipients(ctx context.Context) (int64, error)
	// GetDelivery returns apperrors.ErrNotFound for a customer promotions do
	// not go to (any more).
	GetDelivery(ctx context.Context, promotionID, userID uuid.UUID) (*PromotionDelivery, error)
	// MarkSent records how many customers the promotion was queued for, and
	// reports whether it was still sending.
	MarkSent(ctx context.Context, id uuid.UUID, recipients int32) (bool, error)
}

// CreatePromotionParams is a promotion a manager sends, checked.
type CreatePromotionParams struct {
	Subject   string
	Body      string
	CreatedBy uuid.UUID
}

// ManagerPromotion is a promotion as managers list it. RecipientCount is nil
// while it is sending.
type ManagerPromotion struct {
	ID             uuid.UUID
	Subject        string
	Status         domainpromotion.Status
	RecipientCount *int32
	CreatedAt      time.Time
	SenderName     *string
}

// PromotionDelivery is what one customer's email of a promotion says and
// where it goes.
type PromotionDelivery struct {
	Subject string
	Body    string
	Email   string
	Name    *string
}

// PromotionTask asks for a promotion to be emailed to every customer it goes
// to; the task id is the event's.
type PromotionTask struct {
	EventID     uuid.UUID
	PromotionID uuid.UUID
}

// PromotionEmailTask asks for one customer's email of a promotion. It carries
// ids only: the address and the unsubscribe link are read and made when the
// email is sent.
type PromotionEmailTask struct {
	PromotionID uuid.UUID
	UserID      uuid.UUID
}

// PromotionEmailQueue holds promotions until a worker emails them.
type PromotionEmailQueue interface {
	EnqueuePromotion(ctx context.Context, task PromotionTask) error
	// EnqueuePromotionEmail queues one task per promotion and customer.
	EnqueuePromotionEmail(ctx context.Context, task PromotionEmailTask) error
}

// PromotionEmail is what one customer's email of a promotion says; its
// unsubscribe link carries UnsubscribeToken.
type PromotionEmail struct {
	To               string
	ToName           string
	Subject          string
	Body             string
	UnsubscribeToken string
}

// PromotionEmailComposer writes a promotion email.
type PromotionEmailComposer interface {
	ComposePromotionEmail(msg PromotionEmail) (Email, error)
}
