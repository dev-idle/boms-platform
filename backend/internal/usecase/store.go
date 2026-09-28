package usecase

import (
	"context"
	"slices"
	"strings"
	"time"

	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// StoreUsecase answers what anyone may ask about the bakery's pickup rules and
// free pickup slots.
type StoreUsecase struct {
	store  port.StoreSettingsRepository
	orders port.OrderRepository
}

func NewStoreUsecase(store port.StoreSettingsRepository, orders port.OrderRepository) *StoreUsecase {
	return &StoreUsecase{store: store, orders: orders}
}

// PickupRules returns the rules a checkout is held to right now.
func (u *StoreUsecase) PickupRules(ctx context.Context) (*dto.PickupRulesResponse, error) {
	settings, closed, err := readPickupRules(ctx, u.store, time.Now())
	if err != nil {
		return nil, err
	}
	dates := make([]dto.PublicClosedDateResponse, 0, len(closed))
	for _, c := range closed {
		dates = append(dates, dto.PublicClosedDateResponse{Date: c.Day.Format(domainstore.DayLayout), Reason: c.Reason})
	}
	return &dto.PickupRulesResponse{
		OpensAt:                domainstore.FormatClock(settings.OpensAt),
		ClosesAt:               domainstore.FormatClock(settings.ClosesAt),
		SlotMinutes:            int(settings.SlotLength / time.Minute),
		PreorderMinLeadMinutes: int(settings.PreorderMinLead / time.Minute),
		InstantPrepMinutes:     int(settings.InstantPrep / time.Minute),
		MaxAdvanceDays:         settings.MaxAdvanceDays,
		ClosedDates:            dates,
	}, nil
}

// PickupSlots lists the pickup slots of a bakery day (YYYY-MM-DD) inside the
// booking window, each marked full once it holds as many orders as a slot
// takes. A closed day has none. Which slots suit a cart is the picker's to
// work out from the rules; checkout decides.
func (u *StoreUsecase) PickupSlots(ctx context.Context, date string) (*dto.PickupSlotsResponse, error) {
	day, err := time.Parse(domainstore.DayLayout, strings.TrimSpace(date))
	if err != nil {
		return nil, apperrors.ErrValidation.WithDetail("date", "use YYYY-MM-DD")
	}
	now := time.Now()
	settings, closed, err := readPickupRules(ctx, u.store, now)
	if err != nil {
		return nil, err
	}
	if day.Before(domainstore.DayOf(now)) || day.After(settings.LastPickupDay(now)) {
		return nil, apperrors.ErrValidation.WithDetail("date", "outside the booking window")
	}
	resp := &dto.PickupSlotsResponse{Date: day.Format(domainstore.DayLayout), Slots: []dto.PickupSlotResponse{}}
	if slices.ContainsFunc(closed, func(c domainstore.ClosedDate) bool { return c.Day.Equal(day) }) {
		return resp, nil
	}
	starts := settings.SlotStarts(day)
	due, err := u.orders.CountByPickupTime(ctx, starts[0], starts[len(starts)-1].Add(settings.SlotLength))
	if err != nil {
		return nil, err
	}
	// Orders count against the slot their pickup falls in, keyed by its start.
	booked := make(map[int64]int, len(starts))
	for _, count := range due {
		if start, ok := settings.SlotStartOf(count.At); ok {
			booked[start.Unix()] += count.Orders
		}
	}
	for _, start := range starts {
		resp.Slots = append(resp.Slots, dto.PickupSlotResponse{
			StartsAt: start.UTC().Format(time.RFC3339),
			Full:     booked[start.Unix()] >= settings.SlotCapacity,
		})
	}
	return resp, nil
}
