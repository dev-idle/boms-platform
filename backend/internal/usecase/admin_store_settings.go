package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
)

// AdminStoreSettingsUsecase lets admins change the pickup rules without a
// deploy. Every change is audited and announced to open pages as it commits.
type AdminStoreSettingsUsecase struct {
	store       port.StoreSettingsRepository
	orders      port.OrderRepository
	payment     *PaymentUsecase
	transitions orderTransitions
	tx          port.TxManager
	events      port.EventOutbox
	audit       *auditlogger.Service
	log         *zap.Logger
}

func NewAdminStoreSettingsUsecase(
	store port.StoreSettingsRepository,
	orders port.OrderRepository,
	tickets port.TicketRepository,
	discounts port.DiscountCodeRepository,
	payments port.PaymentRepository,
	payment *PaymentUsecase,
	tx port.TxManager,
	events port.EventOutbox,
	audit *auditlogger.Service,
	log *zap.Logger,
) *AdminStoreSettingsUsecase {
	return &AdminStoreSettingsUsecase{
		store:   store,
		orders:  orders,
		payment: payment,
		transitions: orderTransitions{
			tx: tx, orders: orders, tickets: tickets, discounts: discounts, payments: payments, events: events,
		},
		tx:     tx,
		events: events,
		audit:  audit,
		log:    log,
	}
}

const (
	// maxStrandedListed is how many stranded orders a preview lists; its
	// total counts them all.
	maxStrandedListed = 100
	// strandedBatch is how many stranded orders one pass cancels.
	strandedBatch = 50
)

// StrandedPickups lists the orders a closed day or new opening hours would
// strand, for the admin to see before making the change.
func (u *AdminStoreSettingsUsecase) StrandedPickups(ctx context.Context, q dto.StrandedPickupsQuery) (*dto.StrandedPickupsResponse, error) {
	params, err := u.strandedParams(ctx, q, time.Now())
	if err != nil {
		return nil, err
	}
	rows, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]port.StrandedPickup, error) {
			listed := params
			listed.Limit = maxStrandedListed
			return u.orders.ListStrandedPickups(ctx, listed)
		},
		func(ctx context.Context) (int64, error) {
			return u.orders.CountStrandedPickups(ctx, params)
		},
	)
	if err != nil {
		return nil, err
	}
	out := make([]dto.StrandedPickupResponse, 0, len(rows))
	for _, row := range rows {
		customer := row.CustomerDisplayName
		if customer == nil {
			customer = row.CustomerEmail
		}
		out = append(out, dto.StrandedPickupResponse{
			ID: row.ID.String(), Code: row.Code, Status: string(row.Status), PickupAt: row.PickupAt, Customer: customer,
		})
	}
	return &dto.StrandedPickupsResponse{Orders: out, Total: total}, nil
}

// CancelStrandedPickups cancels the orders a change would strand, each
// refunded in full and its customer emailed the reason, so the day can close
// or the hours change. Each order is tried once, a page at a time; one that
// cannot be cancelled is reported, and the change stays refused while it
// remains.
func (u *AdminStoreSettingsUsecase) CancelStrandedPickups(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	req dto.CancelStrandedPickupsRequest,
) (*dto.CancelStrandedPickupsResponse, error) {
	params, err := u.strandedParams(ctx, req.StrandedPickupsQuery, time.Now())
	if err != nil {
		return nil, err
	}
	reason, err := domainorder.NewCancelReason(req.Reason)
	if err != nil {
		return nil, err
	}
	params.Limit = strandedBatch
	actor := port.OrderActor{ID: actorID, Role: actorRole}
	cancelled := 0
	var failures []error
	for {
		batch, err := u.orders.ListStrandedPickups(ctx, params)
		if err != nil {
			failures = append(failures, err)
			break
		}
		for _, order := range batch {
			done, err := u.cancelStranded(ctx, actor, order, params, reason)
			if err != nil {
				failures = append(failures, apperrors.Errorf("cancel order %s: %w", order.Code, err))
				continue
			}
			if done {
				cancelled++
			}
		}
		if len(batch) < strandedBatch {
			break
		}
		last := batch[len(batch)-1]
		params.AfterPickupAt, params.AfterID = &last.PickupAt, &last.ID
	}
	if len(failures) > 0 {
		return nil, errors.Join(failures...)
	}
	return &dto.CancelStrandedPickupsResponse{Cancelled: cancelled}, nil
}

// cancelStranded cancels one order the change strands, as the bakery: a paid
// one is refunded in full and its customer emailed the reason. An unpaid one is
// first checked with the provider, as a customer's cancellation is, and one
// whose payment the provider is reviewing is left (domainpayment.ErrUnderReview).
// Under the order's lock it is cancelled only while the change still strands
// it — one moved or handed over meanwhile is left, reporting false — and the
// cancellation commits with its audit row.
func (u *AdminStoreSettingsUsecase) cancelStranded(
	ctx context.Context,
	actor port.OrderActor,
	order port.StrandedPickup,
	params port.StrandedPickupsParams,
	reason string,
) (bool, error) {
	if order.Status == domainorder.StatusAwaitingPayment {
		taken, err := u.payment.settle(ctx, order.ID)
		if err != nil {
			return false, err
		}
		if taken == domainpayment.StatusPending {
			return false, domainpayment.ErrUnderReview
		}
	}
	done := false
	err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		// The order as it is now: a payment just settled confirms it.
		locked, err := u.orders.LockForUpdate(txCtx, order.ID)
		if err != nil {
			return err
		}
		if !locked.Status.StillToHandOver() || locked.PickupAt == nil || !strands(params, *locked.PickupAt) {
			return nil
		}
		if _, err := u.transitions.applyInTx(txCtx, &actor, port.UpdateOrderStatusParams{
			OrderID: order.ID, FromStatus: locked.Status, ToStatus: domainorder.StatusCancelled,
		}, reason); err != nil {
			return err
		}
		done = true
		return u.audit.Log(txCtx, domainstore.AuditActionAdminCancelledStrandedOrder, actor.ID, actor.Role, &order.ID, "order",
			map[string]any{"status": locked.Status, "pickup_at": locked.PickupAt},
			map[string]any{"status": domainorder.StatusCancelled, "reason": reason})
	})
	return done && err == nil, err
}

// strands reports whether a pickup at at falls where params find it.
func strands(params port.StrandedPickupsParams, at time.Time) bool {
	if at.Before(params.From) || !at.Before(params.To) {
		return false
	}
	if params.OpensMinute == nil {
		return true
	}
	local := at.In(domainstore.Location)
	minute := utils.Int32FromInt64(int64(local.Hour()*60 + local.Minute()))
	return minute < *params.OpensMinute || minute >= *params.ClosesMinute
}

// strandedParams reads the change q names, checked as the change itself would
// be: a day the bakery may close, or opening hours the settings accept.
func (u *AdminStoreSettingsUsecase) strandedParams(ctx context.Context, q dto.StrandedPickupsQuery, now time.Time) (port.StrandedPickupsParams, error) {
	switch {
	case q.ClosedDate != "" && q.OpensAt == "" && q.ClosesAt == "":
		day, err := time.ParseInLocation(domainstore.DayLayout, q.ClosedDate, domainstore.Location)
		if err != nil {
			return port.StrandedPickupsParams{}, apperrors.ErrValidation.WithDetail("closed_date", "use YYYY-MM-DD")
		}
		if from, to := domainstore.ClosedDateWindow(now); day.Before(from) || day.After(to) {
			return port.StrandedPickupsParams{}, domainstore.ErrClosedDateOutOfRange
		}
		return closedDayParams(day), nil
	case q.ClosedDate == "" && q.OpensAt != "" && q.ClosesAt != "":
		opens, err := domainstore.ParseClock(q.OpensAt)
		if err != nil {
			return port.StrandedPickupsParams{}, apperrors.ErrValidation.WithDetail("opens_at", "use HH:MM")
		}
		closes, err := domainstore.ParseClock(q.ClosesAt)
		if err != nil {
			return port.StrandedPickupsParams{}, apperrors.ErrValidation.WithDetail("closes_at", "use HH:MM")
		}
		settings, err := u.store.GetSettings(ctx)
		if err != nil {
			return port.StrandedPickupsParams{}, err
		}
		settings.OpensAt, settings.ClosesAt = opens, closes
		if err := settings.Validate(); err != nil {
			return port.StrandedPickupsParams{}, err
		}
		return hoursParams(opens, closes, now), nil
	default:
		return port.StrandedPickupsParams{}, apperrors.ErrValidation.WithDetail("closed_date", "name a closed date, or the opening and closing times")
	}
}

// closedDayParams finds the orders on a bakery day.
func closedDayParams(day time.Time) port.StrandedPickupsParams {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, domainstore.Location)
	return port.StrandedPickupsParams{From: start, To: start.AddDate(0, 0, 1)}
}

// hoursParams finds the orders from now on outside the hours [opens, closes),
// as far as any booking window reaches.
func hoursParams(opens, closes time.Duration, now time.Time) port.StrandedPickupsParams {
	opensMinute := utils.Int32FromInt64(int64(opens / time.Minute))
	closesMinute := utils.Int32FromInt64(int64(closes / time.Minute))
	return port.StrandedPickupsParams{
		From: now, To: now.AddDate(0, 0, domainstore.MaxAdvanceDaysLimit+1),
		OpensMinute: &opensMinute, ClosesMinute: &closesMinute,
	}
}

// refuseStranded refuses a change while orders it would strand remain.
func (u *AdminStoreSettingsUsecase) refuseStranded(txCtx context.Context, params port.StrandedPickupsParams) error {
	stranded, err := u.orders.CountStrandedPickups(txCtx, params)
	if err != nil {
		return err
	}
	if stranded > 0 {
		return domainstore.ErrPickupsStranded
	}
	return nil
}

func (u *AdminStoreSettingsUsecase) GetSettings(ctx context.Context) (*dto.StoreSettingsResponse, error) {
	settings, err := u.store.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	return toStoreSettingsResponse(settings), nil
}

// PatchSettings applies the fields req names to the current settings.
func (u *AdminStoreSettingsUsecase) PatchSettings(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	req dto.PatchStoreSettingsRequest,
) (*dto.StoreSettingsResponse, error) {
	if req.OpensAt == nil && req.ClosesAt == nil && req.PreorderMinLeadMinutes == nil && req.MaxAdvanceDays == nil &&
		req.SlotMinutes == nil && req.SlotCapacity == nil && req.InstantPrepMinutes == nil && req.PaymentHoldMinutes == nil {
		return nil, apperrors.ErrValidation.WithDetail("body", "name at least one setting to change")
	}
	var before, after domainstore.Settings
	err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		current, err := u.store.GetSettingsForUpdate(txCtx)
		if err != nil {
			return err
		}
		next, err := patchedSettings(current, req)
		if err != nil {
			return err
		}
		if err := next.Validate(); err != nil {
			return err
		}
		if next.OpensAt != current.OpensAt || next.ClosesAt != current.ClosesAt {
			if err := u.refuseStranded(txCtx, hoursParams(next.OpensAt, next.ClosesAt, time.Now())); err != nil {
				return err
			}
		}
		saved, err := u.store.UpdateSettings(txCtx, next)
		if err != nil {
			return err
		}
		if err := u.events.Add(txCtx, domainstore.SettingsUpdatedEvent()); err != nil {
			return err
		}
		before, after = current, saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := toStoreSettingsResponse(after)
	recordAudit(u.log, u.audit, ctx, domainstore.AuditActionAdminUpdatedSettings, actorID, actorRole, nil, "store_settings",
		toStoreSettingsResponse(before), resp)
	return resp, nil
}

func patchedSettings(current domainstore.Settings, req dto.PatchStoreSettingsRequest) (domainstore.Settings, error) {
	next := current
	if req.OpensAt != nil {
		opens, err := domainstore.ParseClock(*req.OpensAt)
		if err != nil {
			return domainstore.Settings{}, apperrors.ErrValidation.WithDetail("opens_at", "use HH:MM")
		}
		next.OpensAt = opens
	}
	if req.ClosesAt != nil {
		closes, err := domainstore.ParseClock(*req.ClosesAt)
		if err != nil {
			return domainstore.Settings{}, apperrors.ErrValidation.WithDetail("closes_at", "use HH:MM")
		}
		next.ClosesAt = closes
	}
	if req.PreorderMinLeadMinutes != nil {
		// Bounded before the conversion, which would wrap a huge value into range.
		minutes := *req.PreorderMinLeadMinutes
		if minutes < 0 || minutes > int(domainstore.MaxPreorderMinLead/time.Minute) {
			return domainstore.Settings{}, domainstore.ErrInvalidLeadTime
		}
		next.PreorderMinLead = time.Duration(minutes) * time.Minute
	}
	if req.MaxAdvanceDays != nil {
		next.MaxAdvanceDays = *req.MaxAdvanceDays
	}
	if req.SlotMinutes != nil {
		minutes := *req.SlotMinutes
		if minutes < 0 || minutes > int(time.Hour/time.Minute) {
			return domainstore.Settings{}, domainstore.ErrInvalidSlotLength
		}
		next.SlotLength = time.Duration(minutes) * time.Minute
	}
	if req.SlotCapacity != nil {
		next.SlotCapacity = *req.SlotCapacity
	}
	if req.InstantPrepMinutes != nil {
		minutes := *req.InstantPrepMinutes
		if minutes < 0 || minutes > int(domainstore.MaxInstantPrep/time.Minute) {
			return domainstore.Settings{}, domainstore.ErrInvalidInstantPrep
		}
		next.InstantPrep = time.Duration(minutes) * time.Minute
	}
	if req.PaymentHoldMinutes != nil {
		minutes := *req.PaymentHoldMinutes
		if minutes < 0 || minutes > int(domainstore.MaxPaymentHold/time.Minute) {
			return domainstore.Settings{}, domainstore.ErrInvalidPaymentHold
		}
		next.PaymentHold = time.Duration(minutes) * time.Minute
	}
	return next, nil
}

// ListClosedDates returns the closures from today on, earliest first.
func (u *AdminStoreSettingsUsecase) ListClosedDates(ctx context.Context) ([]dto.ClosedDateResponse, error) {
	from, to := domainstore.ClosedDateWindow(time.Now())
	closed, err := u.store.ListClosedDates(ctx, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ClosedDateResponse, 0, len(closed))
	for _, c := range closed {
		out = append(out, toClosedDateResponse(c))
	}
	return out, nil
}

// AddClosedDate closes the bakery on a day from today to a year ahead.
func (u *AdminStoreSettingsUsecase) AddClosedDate(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	req dto.CreateClosedDateRequest,
) (*dto.ClosedDateResponse, error) {
	day, err := time.Parse(domainstore.DayLayout, req.Date)
	if err != nil {
		return nil, apperrors.ErrValidation.WithDetail("date", "use YYYY-MM-DD")
	}
	reason, err := domainstore.NewClosedDate(day, req.Reason, time.Now())
	if err != nil {
		return nil, err
	}
	var added domainstore.ClosedDate
	err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
		// Bookings hold the settings while they book: taking them for update
		// waits for those in flight, so the count below sees them.
		if _, err := u.store.GetSettingsForUpdate(txCtx); err != nil {
			return err
		}
		if err := u.refuseStranded(txCtx, closedDayParams(day)); err != nil {
			return err
		}
		created, err := u.store.AddClosedDate(txCtx, day, reason)
		if err != nil {
			if errors.Is(err, apperrors.ErrConflict) {
				return domainstore.ErrClosedDateExists
			}
			return err
		}
		added = created
		return u.events.Add(txCtx, domainstore.SettingsUpdatedEvent())
	})
	if err != nil {
		return nil, err
	}
	resp := toClosedDateResponse(added)
	recordAudit(u.log, u.audit, ctx, domainstore.AuditActionAdminAddedClosedDate, actorID, actorRole, &added.ID, "store_closed_date", nil, resp)
	return &resp, nil
}

// RemoveClosedDate reopens the bakery on a closed day.
func (u *AdminStoreSettingsUsecase) RemoveClosedDate(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	id uuid.UUID,
) error {
	var removed domainstore.ClosedDate
	err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		deleted, err := u.store.RemoveClosedDate(txCtx, id)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return domainstore.ErrClosedDateNotFound
			}
			return err
		}
		removed = deleted
		return u.events.Add(txCtx, domainstore.SettingsUpdatedEvent())
	})
	if err != nil {
		return err
	}
	recordAudit(u.log, u.audit, ctx, domainstore.AuditActionAdminRemovedClosedDate, actorID, actorRole, &removed.ID, "store_closed_date",
		toClosedDateResponse(removed), nil)
	return nil
}

func toStoreSettingsResponse(s domainstore.Settings) *dto.StoreSettingsResponse {
	return &dto.StoreSettingsResponse{
		OpensAt:                domainstore.FormatClock(s.OpensAt),
		ClosesAt:               domainstore.FormatClock(s.ClosesAt),
		PreorderMinLeadMinutes: int(s.PreorderMinLead / time.Minute),
		MaxAdvanceDays:         s.MaxAdvanceDays,
		SlotMinutes:            int(s.SlotLength / time.Minute),
		SlotCapacity:           s.SlotCapacity,
		InstantPrepMinutes:     int(s.InstantPrep / time.Minute),
		PaymentHoldMinutes:     int(s.PaymentHold / time.Minute),
		UpdatedAt:              s.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toClosedDateResponse(c domainstore.ClosedDate) dto.ClosedDateResponse {
	return dto.ClosedDateResponse{
		ID:        c.ID.String(),
		Date:      c.Day.Format(domainstore.DayLayout),
		Reason:    c.Reason,
		CreatedAt: c.CreatedAt.UTC().Format(time.RFC3339),
	}
}
