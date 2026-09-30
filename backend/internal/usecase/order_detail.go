package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/shared/ctxmeta"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// orderDetailParts is what an order's detail shows besides the order itself.
type orderDetailParts struct {
	items    []domainorder.Item
	timeline []domainorder.StatusEvent
	tickets  []domainorder.Ticket
	payment  *domainpayment.Payment
}

// readOrderDetail reads an order's lines, its status history, its tickets and
// its payment at once, so a detail page pays one round trip for all four. Like
// listWithTotal it refuses to run inside a transaction, whose single
// connection cannot serve concurrent reads.
func readOrderDetail(
	ctx context.Context,
	orders port.OrderRepository,
	tickets port.TicketRepository,
	payments port.PaymentRepository,
	orderID uuid.UUID,
) (orderDetailParts, error) {
	if ctxmeta.InTransaction(ctx) {
		return orderDetailParts{}, apperrors.Errorf("read order detail: cannot run inside a transaction")
	}
	var parts orderDetailParts
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		found, err := orders.ListItemsByOrderID(groupCtx, orderID)
		parts.items = found
		return err
	})
	group.Go(func() error {
		found, err := orders.ListStatusEvents(groupCtx, orderID)
		parts.timeline = found
		return err
	})
	group.Go(func() error {
		found, err := tickets.ListByOrder(groupCtx, orderID)
		parts.tickets = found
		return err
	})
	group.Go(func() error {
		found, err := payments.GetByOrderID(groupCtx, orderID)
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		parts.payment = found
		return err
	})
	if err := group.Wait(); err != nil {
		return orderDetailParts{}, err
	}
	return parts, nil
}
