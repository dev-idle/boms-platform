package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// PaymentRepository keeps orders' payments.
type PaymentRepository struct {
	queries *sqlcgen.Queries
}

func NewPaymentRepository(pool *Pool) *PaymentRepository {
	return &PaymentRepository{queries: pool.Queries()}
}

func (r *PaymentRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// Create implements port.PaymentRepository.
func (r *PaymentRepository) Create(ctx context.Context, params port.CreatePaymentParams) (*domainpayment.Payment, error) {
	row, err := r.q(ctx).CreatePayment(ctx, sqlcgen.CreatePaymentParams{
		OrderID:         params.OrderID,
		Provider:        sqlcgen.PaymentProvider(params.Provider),
		ProviderOrderID: params.ProviderOrderID,
		ApproveUrl:      params.ApproveURL,
		AmountCents:     params.AmountCents,
		Currency:        params.Currency,
	})
	if err != nil {
		err = mapRepoError(err, "create payment")
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrConflict
		}
		return nil, err
	}
	return mapPayment(row), nil
}

// GetByOrderID implements port.PaymentRepository.
func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domainpayment.Payment, error) {
	row, err := r.q(ctx).GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		return nil, mapRepoError(err, "get payment by order")
	}
	return mapPayment(row), nil
}

// GetByProviderOrderID implements port.PaymentRepository.
func (r *PaymentRepository) GetByProviderOrderID(
	ctx context.Context,
	provider domainpayment.Provider,
	providerOrderID string,
) (*domainpayment.Payment, error) {
	row, err := r.q(ctx).GetPaymentByProviderOrderID(ctx, sqlcgen.GetPaymentByProviderOrderIDParams{
		Provider:        sqlcgen.PaymentProvider(provider),
		ProviderOrderID: providerOrderID,
	})
	if err != nil {
		return nil, mapRepoError(err, "get payment by provider order")
	}
	return mapPayment(row), nil
}

// Restart implements port.PaymentRepository.
func (r *PaymentRepository) Restart(ctx context.Context, orderID uuid.UUID, providerOrderID, approveURL string) (*domainpayment.Payment, error) {
	row, err := r.q(ctx).RestartPayment(ctx, sqlcgen.RestartPaymentParams{
		OrderID:         orderID,
		ProviderOrderID: providerOrderID,
		ApproveUrl:      approveURL,
	})
	if err != nil {
		return nil, mapRepoError(err, "restart payment")
	}
	return mapPayment(row), nil
}

// RecordCapture implements port.PaymentRepository.
func (r *PaymentRepository) RecordCapture(ctx context.Context, paymentID uuid.UUID, capture domainpayment.Capture) (*domainpayment.Payment, error) {
	row, err := r.q(ctx).RecordPaymentCapture(ctx, sqlcgen.RecordPaymentCaptureParams{
		ID:        paymentID,
		Status:    sqlcgen.PaymentStatus(capture.Status),
		CaptureID: &capture.ID,
	})
	if err != nil {
		return nil, mapRepoError(err, "record payment capture")
	}
	return mapPayment(row), nil
}

func mapPayment(row sqlcgen.Payment) *domainpayment.Payment {
	return &domainpayment.Payment{
		ID:              row.ID,
		OrderID:         row.OrderID,
		Provider:        domainpayment.Provider(row.Provider),
		ProviderOrderID: row.ProviderOrderID,
		ApproveURL:      row.ApproveUrl,
		Status:          domainpayment.Status(row.Status),
		CaptureID:       row.CaptureID,
		AmountCents:     row.AmountCents,
		Currency:        row.Currency,
		CapturedAt:      row.CapturedAt,
		CreatedAt:       row.CreatedAt,
	}
}

var _ port.PaymentRepository = (*PaymentRepository)(nil)
