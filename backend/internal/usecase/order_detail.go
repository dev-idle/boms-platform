package usecase

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/shared/ctxmeta"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// orderLinesAndTimeline reads an order's lines and its status history at once,
// so a detail page pays one round trip for both. Like listWithTotal it refuses
// to run inside a transaction, whose single connection cannot serve two reads.
func orderLinesAndTimeline(
	ctx context.Context,
	orders port.OrderRepository,
	orderID uuid.UUID,
) ([]domainorder.Item, []domainorder.StatusEvent, error) {
	if ctxmeta.InTransaction(ctx) {
		return nil, nil, apperrors.Errorf("read order detail: cannot run inside a transaction")
	}
	var (
		items    []domainorder.Item
		timeline []domainorder.StatusEvent
	)
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		found, err := orders.ListItemsByOrderID(groupCtx, orderID)
		items = found
		return err
	})
	group.Go(func() error {
		found, err := orders.ListStatusEvents(groupCtx, orderID)
		timeline = found
		return err
	})
	if err := group.Wait(); err != nil {
		return nil, nil, err
	}
	return items, timeline, nil
}
