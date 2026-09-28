package usecase

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/shared/ctxmeta"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// readPickupRules reads the settings and the closed days inside their booking
// window, checked at now. The widest window an admin may set is known before
// the settings are, so both queries run at once and a checkout pays one round
// trip for them. Like listWithTotal it refuses to run inside a transaction.
func readPickupRules(
	ctx context.Context,
	repo port.StoreSettingsRepository,
	now time.Time,
) (domainstore.Settings, []domainstore.ClosedDate, error) {
	if ctxmeta.InTransaction(ctx) {
		return domainstore.Settings{}, nil, apperrors.Errorf("read pickup rules: cannot run inside a transaction")
	}
	today := domainstore.DayOf(now)
	var (
		settings domainstore.Settings
		closed   []domainstore.ClosedDate
	)
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		found, err := repo.GetSettings(groupCtx)
		settings = found
		return err
	})
	group.Go(func() error {
		found, err := repo.ListClosedDates(groupCtx, today, today.AddDate(0, 0, domainstore.MaxAdvanceDaysLimit))
		closed = found
		return err
	})
	if err := group.Wait(); err != nil {
		return domainstore.Settings{}, nil, err
	}

	last := domainstore.DayOf(now.Add(time.Duration(settings.MaxAdvanceDays) * 24 * time.Hour))
	inWindow := closed[:0]
	for _, c := range closed {
		if !c.Day.After(last) {
			inWindow = append(inWindow, c)
		}
	}
	return settings, inWindow, nil
}

func pickupPolicy(settings domainstore.Settings, closed []domainstore.ClosedDate) domainorder.PickupPolicy {
	days := make([]time.Time, 0, len(closed))
	for _, c := range closed {
		days = append(days, c.Day)
	}
	return domainorder.PickupPolicy{Settings: settings, ClosedDays: days}
}
