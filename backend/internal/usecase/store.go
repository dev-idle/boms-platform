package usecase

import (
	"context"
	"time"

	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
)

// StoreUsecase answers what anyone may ask about the bakery's pickup rules.
type StoreUsecase struct {
	store port.StoreSettingsRepository
}

func NewStoreUsecase(store port.StoreSettingsRepository) *StoreUsecase {
	return &StoreUsecase{store: store}
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
		PreorderMinLeadMinutes: int(settings.PreorderMinLead / time.Minute),
		MaxAdvanceDays:         settings.MaxAdvanceDays,
		ClosedDates:            dates,
	}, nil
}
