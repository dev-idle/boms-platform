package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
)

// CreateOrder takes an order at the counter or on the phone, for a
// customer's account or for a guest: priced from the catalog now, held to the
// pickup rules a checkout is, and confirmed at once with its total due in cash
// when it is collected. checkoutKey is the request's Idempotency-Key: the same
// key returns the order it took, so a retry takes nothing twice.
func (u *StaffOrderUsecase) CreateOrder(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	checkoutKey uuid.UUID,
	req dto.CreateStaffOrderRequest,
) (*dto.StaffOrderResponse, error) {
	if taken, err := u.orders.GetByStaffCheckoutKey(ctx, checkoutKey); err == nil {
		return u.Get(ctx, taken.ID)
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	channel := domainorder.Channel(req.Channel)
	if !channel.TakenByStaff() {
		return nil, apperrors.ErrValidation.WithDetail("channel", "counter or phone")
	}
	customerID, guest, err := staffOrderCustomer(req)
	if err != nil {
		return nil, err
	}
	items, err := staffOrderItems(req.Items)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	settings, closed, err := readPickupRules(ctx, u.store, now)
	if err != nil {
		return nil, err
	}
	policy := pickupPolicy(settings, closed)
	actor := &port.OrderActor{ID: actorID, Role: actorRole}

	var created *domainorder.Order
	err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
		if customerID != nil {
			// Held until the order commits, as a checkout holds its customer:
			// closing the account waits for the order.
			customer, err := u.users.GetByIDForShare(txCtx, *customerID)
			if errors.Is(err, apperrors.ErrNotFound) || (err == nil && customer.Role != domainuser.RoleCustomer) {
				return apperrors.ErrValidation.WithDetail("customer_id", "no open customer account")
			}
			if err != nil {
				return err
			}
		}
		lines, needs, err := u.pricedItems(txCtx, items)
		if err != nil {
			return err
		}
		totals := summarizeCart(lines, nil, now.UTC())
		booking, err := bookPickup(txCtx, u.orders, customerID, needs, req.PickupAt, now, policy, nil)
		if err != nil {
			return err
		}
		// Taken last before the insert: the day's number stays locked until commit.
		day, number, err := u.orders.NextDayNumber(txCtx)
		if err != nil {
			return err
		}
		order, err := u.orders.Create(txCtx, port.CreateOrderParams{
			UserID:        customerID,
			Guest:         guest,
			Channel:       channel,
			Code:          domainorder.Code(day, number),
			Status:        domainorder.StatusConfirmed,
			Type:          booking.orderType,
			SubtotalCents: totals.SubtotalCents,
			TotalCents:    totals.SubtotalCents,
			PickupAt:      &req.PickupAt,
			CheckoutKey:   &checkoutKey,
		})
		if err != nil {
			return err
		}
		if err := u.orders.CreateItems(txCtx, orderItemsOf(order.ID, lines)); err != nil {
			return err
		}
		if err := createOrderTickets(txCtx, u.tickets, order.ID); err != nil {
			return err
		}
		if err := u.orders.AddStatusEvent(txCtx, port.AddOrderStatusEventParams{
			OrderID: order.ID,
			To:      order.Status,
			Actor:   actor,
		}); err != nil {
			return err
		}
		if order.TotalCents > 0 {
			if err := u.payments.CreateCash(txCtx, order.ID, order.TotalCents); err != nil {
				return err
			}
		}
		if err := u.events.Add(txCtx, domainorder.CreatedEvent(*order)); err != nil {
			return err
		}
		if booking.fillsSlot {
			if err := u.events.Add(txCtx, domainorder.SlotsChangedEvent(domainstore.DayOf(req.PickupAt))); err != nil {
				return err
			}
		}
		created = order
		return nil
	})
	if err != nil {
		// A request with the same key may have taken the order while this one
		// waited, and then found its slot full or the key taken.
		if taken, lookupErr := u.orders.GetByStaffCheckoutKey(ctx, checkoutKey); lookupErr == nil {
			return u.Get(ctx, taken.ID)
		}
		return nil, err
	}
	return u.Get(ctx, created.ID)
}

// pricedItems prices the items of an order staff take from the catalog now,
// and says what they ask of the bakery. A product gone from the catalog, or a
// custom cake, which is configured online, is refused.
func (u *StaffOrderUsecase) pricedItems(ctx context.Context, items []domaincart.Item) ([]pricedCartLine, domainorder.Fulfillment, error) {
	lines, err := u.cartUC.pricer.priceLines(ctx, items)
	if err != nil {
		return nil, domainorder.Fulfillment{}, err
	}
	for _, line := range lines {
		if !line.IsAvailable {
			return nil, domainorder.Fulfillment{}, domaincart.ErrProductUnavailable
		}
	}
	needs, err := u.cartUC.fulfillmentOf(ctx, lines)
	return lines, needs, err
}

// Quote prices the items of an order staff are about to take and says what
// they ask of the bakery, so the counter can tell the total and offer a
// pickup that suits them; taking the order prices it again.
func (u *StaffOrderUsecase) Quote(ctx context.Context, req dto.StaffOrderQuoteRequest) (*dto.StaffOrderQuoteResponse, error) {
	items, err := staffOrderItems(req.Items)
	if err != nil {
		return nil, err
	}
	lines, needs, err := u.pricedItems(ctx, items)
	if err != nil {
		return nil, err
	}
	return &dto.StaffOrderQuoteResponse{
		TotalCents:  summarizeCart(lines, nil, time.Now().UTC()).SubtotalCents,
		Fulfillment: mapFulfillmentToDTO(needs),
	}, nil
}

// FindCustomer is the open customer account with that email, for staff to
// take an order for.
func (u *StaffOrderUsecase) FindCustomer(ctx context.Context, email string) (*dto.StaffCustomerResponse, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, apperrors.ErrValidation.WithDetail("email", "an email address")
	}
	customer, err := u.users.FindCustomerByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &dto.StaffCustomerResponse{
		ID:          customer.ID.String(),
		Email:       customer.Email,
		DisplayName: customer.DisplayName,
		Phone:       customer.Phone,
	}, nil
}

// staffOrderCustomer is who an order staff take is for: a customer's account,
// or a guest with a name and a Vietnam mobile number — exactly one of them.
func staffOrderCustomer(req dto.CreateStaffOrderRequest) (*uuid.UUID, *domainorder.Guest, error) {
	if (req.CustomerID == nil) == (req.Guest == nil) {
		return nil, nil, apperrors.ErrValidation.WithDetail("customer_id", "a customer account or a guest, not both")
	}
	if req.CustomerID != nil {
		id, err := uuid.Parse(*req.CustomerID)
		if err != nil {
			return nil, nil, apperrors.ErrValidation.WithDetail("customer_id", "invalid uuid")
		}
		return &id, nil, nil
	}
	phone, ok := utils.NormalizeVietnamPhone(strings.TrimSpace(req.Guest.Phone))
	if !ok {
		return nil, nil, apperrors.ErrValidation.WithDetail("guest.phone", "a Vietnam mobile number")
	}
	guest, err := domainorder.NewGuest(req.Guest.Name, phone)
	if err != nil {
		return nil, nil, err
	}
	return nil, &guest, nil
}

// staffOrderItems reads the lines of an order staff take, each product or
// combo once.
func staffOrderItems(reqs []dto.StaffOrderItemRequest) ([]domaincart.Item, error) {
	items := make([]domaincart.Item, 0, len(reqs))
	seen := make(map[uuid.UUID]bool, len(reqs))
	for _, req := range reqs {
		productID, comboID, err := parseCartLineTarget(req.ProductID, req.ComboID)
		if err != nil {
			return nil, err
		}
		item := domaincart.Item{Quantity: req.Quantity, ProductID: productID, ComboID: comboID}
		target := comboID
		item.LineType = domaincart.LineTypeCombo
		if productID != nil {
			target = productID
			item.LineType = domaincart.LineTypeProduct
		}
		if seen[*target] {
			return nil, apperrors.ErrValidation.WithDetail("items", "each product or combo once")
		}
		seen[*target] = true
		items = append(items, item)
	}
	return items, nil
}
