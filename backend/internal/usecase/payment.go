package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// approvedQuery marks the buyer's return from the provider after approving:
// the order page captures the payment when it sees it.
const approvedQuery = "?paypal=approved"

const (
	// expiryGrace is how long past its due time an order waits before it
	// expires, so a capture started in time finishes first: well past the
	// longest the provider's answer can take.
	expiryGrace = 2 * time.Minute
	// expiryBatch is how many overdue orders one run expires; the next run
	// takes the rest.
	expiryBatch = 50
)

// PaymentUsecase takes an order's payment through the provider, confirms the
// order once the money is taken, and expires the order that is not paid in
// time.
type PaymentUsecase struct {
	tx          port.TxManager
	orders      port.OrderRepository
	discounts   port.DiscountCodeRepository
	payments    port.PaymentRepository
	gateway     port.PaymentGateway
	transitions orderTransitions
	siteURL     string
	log         *zap.Logger
}

func NewPaymentUsecase(
	tx port.TxManager,
	orders port.OrderRepository,
	discounts port.DiscountCodeRepository,
	tickets port.TicketRepository,
	payments port.PaymentRepository,
	gateway port.PaymentGateway,
	events port.EventOutbox,
	siteURL string,
	log *zap.Logger,
) *PaymentUsecase {
	return &PaymentUsecase{
		tx:          tx,
		orders:      orders,
		discounts:   discounts,
		payments:    payments,
		gateway:     gateway,
		transitions: orderTransitions{tx: tx, orders: orders, tickets: tickets, events: events},
		siteURL:     siteURL,
		log:         log,
	}
}

// Start returns the provider page the customer approves their order's payment
// on. A payment already started keeps its page; one the provider denied
// starts over.
func (u *PaymentUsecase) Start(ctx context.Context, userID, orderID uuid.UUID) (string, error) {
	order, err := u.orders.GetByIDForUser(ctx, userID, orderID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return "", domainorder.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if !order.Payable(time.Now()) {
		return "", domainpayment.ErrNotPayable
	}
	current, err := u.payments.GetByOrderID(ctx, orderID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return "", err
	}
	if current != nil && current.Status != domainpayment.StatusDenied {
		return approvePage(current)
	}

	page := u.siteURL + domainorder.PagePath(order.ID)
	providerOrderID, approveURL, err := u.gateway.CreateOrder(ctx, port.PaymentOrderRequest{
		OrderID:     order.ID,
		AmountCents: order.TotalCents,
		Currency:    domainpayment.Currency,
		ReturnURL:   page + approvedQuery,
		CancelURL:   page,
	})
	if err != nil {
		return "", err
	}
	if current == nil {
		_, err = u.payments.Create(ctx, port.CreatePaymentParams{
			OrderID:         order.ID,
			Provider:        domainpayment.ProviderPayPal,
			ProviderOrderID: providerOrderID,
			ApproveURL:      approveURL,
			AmountCents:     order.TotalCents,
			Currency:        domainpayment.Currency,
		})
	} else {
		_, err = u.payments.Restart(ctx, order.ID, providerOrderID, approveURL)
	}
	// Another request for the order got there first: its provider order
	// stands, and the one just made is left unapproved to lapse.
	if errors.Is(err, apperrors.ErrConflict) || errors.Is(err, apperrors.ErrNotFound) {
		if current, err = u.payments.GetByOrderID(ctx, orderID); err != nil {
			return "", err
		}
		return approvePage(current)
	}
	if err != nil {
		return "", err
	}
	return approveURL, nil
}

// approvePage is where the buyer approves a payment not taken yet. A pending
// or taken one is past approving.
func approvePage(p *domainpayment.Payment) (string, error) {
	if p.Status != domainpayment.StatusCreated {
		return "", domainpayment.ErrNotPayable
	}
	return p.ApproveURL, nil
}

// Capture takes the payment the customer approved on the provider's page and
// confirms the order once the money is taken. It answers where the payment
// stands: taken, or held by the provider for review. Asking again, as a
// reloaded return page does, answers the same.
func (u *PaymentUsecase) Capture(ctx context.Context, userID, orderID uuid.UUID) (domainpayment.Status, error) {
	order, err := u.orders.GetByIDForUser(ctx, userID, orderID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return "", domainorder.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	p, err := u.payments.GetByOrderID(ctx, orderID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return "", domainpayment.ErrNotPayable
	}
	if err != nil {
		return "", err
	}
	switch p.Status {
	case domainpayment.StatusCaptured, domainpayment.StatusPending:
		return p.Status, nil
	case domainpayment.StatusDenied:
		return "", domainpayment.ErrNotCompleted
	}
	// Only in time: the order cannot expire before a capture started now ends.
	if !order.Payable(time.Now()) {
		return "", domainpayment.ErrNotPayable
	}
	capture, err := u.gateway.Capture(ctx, p.ProviderOrderID)
	if err != nil {
		return "", err
	}
	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		return u.record(txCtx, p, capture)
	}); err != nil {
		return "", err
	}
	if capture.Status == domainpayment.StatusDenied {
		// Refused: paying again starts a new PayPal order.
		return "", domainpayment.ErrNotCompleted
	}
	return capture.Status, nil
}

// HandleWebhook applies the provider's signed notice about a capture: one held
// for review that went through confirms the order, one refused lets the
// customer pay again. A notice delivered again changes nothing.
func (u *PaymentUsecase) HandleWebhook(ctx context.Context, header func(name string) string, body []byte) error {
	event, err := u.gateway.VerifyWebhook(ctx, header, body)
	if err != nil || event.Capture == nil {
		return err
	}
	p, err := u.payments.GetByProviderOrderID(ctx, domainpayment.ProviderPayPal, event.ProviderOrderID)
	if errors.Is(err, apperrors.ErrNotFound) {
		// Not an order of this shop's, or a provider order a restart replaced.
		u.log.Info("payment_webhook_ignored", zap.String("event_id", event.ID))
		return nil
	}
	if err != nil {
		return err
	}
	return u.tx.WithTx(ctx, func(txCtx context.Context) error {
		return u.record(txCtx, p, *event.Capture)
	})
}

// ExpireOverdue expires the orders whose payment is overdue past the grace, a
// batch at a time, and returns how many expired. Before one expires, the provider is
// asked whether the buyer paid after all: money taken confirms the order, and
// a capture still under review keeps it waiting. Each order stands alone: one
// that fails is reported with the rest while the others go on.
func (u *PaymentUsecase) ExpireOverdue(ctx context.Context) (int, error) {
	ids, err := u.orders.ListDueUnpaid(ctx, expiryGrace, expiryBatch)
	if err != nil {
		return 0, err
	}
	expired := 0
	var failures []error
	for _, id := range ids {
		done, err := u.expire(ctx, id)
		if err != nil {
			failures = append(failures, apperrors.Errorf("expire order %s: %w", id, err))
			continue
		}
		if done {
			expired++
		}
	}
	return expired, errors.Join(failures...)
}

// expire expires one overdue order unless the provider took, or is
// reviewing, its payment; it reports whether the order expired.
func (u *PaymentUsecase) expire(ctx context.Context, orderID uuid.UUID) (bool, error) {
	p, err := u.payments.GetByOrderID(ctx, orderID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return false, err
	}
	if p != nil && (p.Status == domainpayment.StatusCreated || p.Status == domainpayment.StatusPending) {
		capture, err := u.gateway.Lookup(ctx, p.ProviderOrderID)
		if err != nil {
			return false, err
		}
		if capture != nil {
			// An answer the return page and the webhook did not record.
			if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
				return u.record(txCtx, p, *capture)
			}); err != nil {
				return false, err
			}
			if capture.Status != domainpayment.StatusDenied {
				return false, nil
			}
		}
	}
	err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
		order, err := u.orders.Expire(txCtx, orderID, expiryGrace)
		if err != nil {
			return err
		}
		if err := u.transitions.recordInTx(txCtx, nil, domainorder.StatusAwaitingPayment, *order); err != nil {
			return err
		}
		if order.DiscountCodeID != nil {
			return u.discounts.ReleaseUse(txCtx, *order.DiscountCodeID)
		}
		return nil
	})
	if errors.Is(err, apperrors.ErrNotFound) {
		// Paid since it was found overdue.
		return false, nil
	}
	return err == nil, err
}

// record keeps the provider's answer to a capture and, when the money is
// taken, confirms the order in the same transaction: the two exist together
// or not at all.
func (u *PaymentUsecase) record(txCtx context.Context, p *domainpayment.Payment, capture domainpayment.Capture) error {
	if capture.Currency != p.Currency || capture.AmountCents != p.AmountCents {
		return apperrors.Errorf("payment %s: captured %d %s for an order of %d %s: %w",
			p.ID, capture.AmountCents, capture.Currency, p.AmountCents, p.Currency, domainpayment.ErrAmountMismatch)
	}
	if _, err := u.payments.RecordCapture(txCtx, p.ID, capture); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return err
	}
	if capture.Status != domainpayment.StatusCaptured {
		return nil
	}
	order, err := u.orders.LockForUpdate(txCtx, p.OrderID)
	if err != nil {
		return err
	}
	customer := &port.OrderActor{ID: order.UserID, Role: domainuser.RoleCustomer}
	_, err = u.transitions.applyInTx(txCtx, customer, port.UpdateOrderStatusParams{
		OrderID:    order.ID,
		FromStatus: domainorder.StatusAwaitingPayment,
		ToStatus:   domainorder.StatusConfirmed,
	})
	if errors.Is(err, domainorder.ErrInvalidStatusTransition) {
		// The order closed while the buyer paid: the capture is kept on record
		// so the money can be returned.
		u.log.Error("payment_captured_for_closed_order",
			zap.String("order_id", order.ID.String()), zap.String("status", string(order.Status)))
		return nil
	}
	return err
}
