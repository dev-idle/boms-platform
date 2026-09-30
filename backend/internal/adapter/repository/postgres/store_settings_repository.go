package postgres

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// maxClosedDateRows bounds a closed-date list: one row per day of a year and a day.
const maxClosedDateRows = 367

type StoreSettingsRepository struct {
	queries *sqlcgen.Queries
}

func NewStoreSettingsRepository(pool *Pool) *StoreSettingsRepository {
	return &StoreSettingsRepository{queries: pool.Queries()}
}

func (r *StoreSettingsRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *StoreSettingsRepository) GetSettings(ctx context.Context) (domainstore.Settings, error) {
	row, err := r.q(ctx).GetStoreSettings(ctx)
	if err != nil {
		return domainstore.Settings{}, mapRepoError(err, "get store settings")
	}
	return settingsFromRow(row), nil
}

func (r *StoreSettingsRepository) GetSettingsForUpdate(ctx context.Context) (domainstore.Settings, error) {
	if txFromContext(ctx) == nil {
		return domainstore.Settings{}, apperrors.Errorf("lock store settings: requires a transaction")
	}
	row, err := r.q(ctx).GetStoreSettingsForUpdate(ctx)
	if err != nil {
		return domainstore.Settings{}, mapRepoError(err, "lock store settings")
	}
	return settingsFromRow(sqlcgen.GetStoreSettingsRow(row)), nil
}

func (r *StoreSettingsRepository) UpdateSettings(ctx context.Context, s domainstore.Settings) (domainstore.Settings, error) {
	params, err := settingsParams(s)
	if err != nil {
		return domainstore.Settings{}, err
	}
	row, err := r.q(ctx).UpdateStoreSettings(ctx, params)
	if err != nil {
		return domainstore.Settings{}, mapRepoError(err, "update store settings")
	}
	return settingsFromRow(sqlcgen.GetStoreSettingsRow(row)), nil
}

func (r *StoreSettingsRepository) ListClosedDates(ctx context.Context, from, to time.Time) ([]domainstore.ClosedDate, error) {
	rows, err := r.q(ctx).ListStoreClosedDates(ctx, sqlcgen.ListStoreClosedDatesParams{
		FromDay: from,
		ToDay:   to,
		MaxRows: maxClosedDateRows,
	})
	if err != nil {
		return nil, mapRepoError(err, "list closed dates")
	}
	out := make([]domainstore.ClosedDate, 0, len(rows))
	for _, row := range rows {
		out = append(out, domainstore.ClosedDate{ID: row.ID, Day: row.ClosedOn, Reason: row.Reason, CreatedAt: row.CreatedAt})
	}
	return out, nil
}

func (r *StoreSettingsRepository) AddClosedDate(ctx context.Context, day time.Time, reason string) (domainstore.ClosedDate, error) {
	row, err := r.q(ctx).CreateStoreClosedDate(ctx, sqlcgen.CreateStoreClosedDateParams{ClosedOn: day, Reason: reason})
	if err != nil {
		return domainstore.ClosedDate{}, mapRepoError(err, "add closed date")
	}
	return domainstore.ClosedDate{ID: row.ID, Day: row.ClosedOn, Reason: row.Reason, CreatedAt: row.CreatedAt}, nil
}

func (r *StoreSettingsRepository) RemoveClosedDate(ctx context.Context, id uuid.UUID) (domainstore.ClosedDate, error) {
	row, err := r.q(ctx).SoftDeleteStoreClosedDate(ctx, id)
	if err != nil {
		return domainstore.ClosedDate{}, mapRepoError(err, "remove closed date")
	}
	return domainstore.ClosedDate{ID: row.ID, Day: row.ClosedOn, Reason: row.Reason, CreatedAt: row.CreatedAt}, nil
}

// settingsFromRow reads the settings row; the lock and update queries return
// the same columns, so their rows convert to this one.
func settingsFromRow(row sqlcgen.GetStoreSettingsRow) domainstore.Settings {
	return domainstore.Settings{
		OpensAt:         time.Duration(row.OpensAtMinute) * time.Minute,
		ClosesAt:        time.Duration(row.ClosesAtMinute) * time.Minute,
		PreorderMinLead: time.Duration(row.PreorderMinLeadMinutes) * time.Minute,
		MaxAdvanceDays:  int(row.MaxAdvanceDays),
		SlotLength:      time.Duration(row.SlotMinutes) * time.Minute,
		SlotCapacity:    int(row.SlotCapacity),
		InstantPrep:     time.Duration(row.InstantPrepMinutes) * time.Minute,
		PaymentHold:     time.Duration(row.PaymentHoldMinutes) * time.Minute,
		UpdatedAt:       row.UpdatedAt,
	}
}

func settingsParams(s domainstore.Settings) (sqlcgen.UpdateStoreSettingsParams, error) {
	opens, err := int16Of("opens_at_minute", int64(s.OpensAt/time.Minute))
	if err != nil {
		return sqlcgen.UpdateStoreSettingsParams{}, err
	}
	closes, err := int16Of("closes_at_minute", int64(s.ClosesAt/time.Minute))
	if err != nil {
		return sqlcgen.UpdateStoreSettingsParams{}, err
	}
	advance, err := int16Of("max_advance_days", int64(s.MaxAdvanceDays))
	if err != nil {
		return sqlcgen.UpdateStoreSettingsParams{}, err
	}
	slot, err := int16Of("slot_minutes", int64(s.SlotLength/time.Minute))
	if err != nil {
		return sqlcgen.UpdateStoreSettingsParams{}, err
	}
	capacity, err := int16Of("slot_capacity", int64(s.SlotCapacity))
	if err != nil {
		return sqlcgen.UpdateStoreSettingsParams{}, err
	}
	prep, err := int16Of("instant_prep_minutes", int64(s.InstantPrep/time.Minute))
	if err != nil {
		return sqlcgen.UpdateStoreSettingsParams{}, err
	}
	hold, err := int16Of("payment_hold_minutes", int64(s.PaymentHold/time.Minute))
	if err != nil {
		return sqlcgen.UpdateStoreSettingsParams{}, err
	}
	lead := int64(s.PreorderMinLead / time.Minute)
	if lead < math.MinInt32 || lead > math.MaxInt32 {
		return sqlcgen.UpdateStoreSettingsParams{}, apperrors.Errorf("preorder_min_lead_minutes %d out of range", lead)
	}
	return sqlcgen.UpdateStoreSettingsParams{
		OpensAtMinute:          opens,
		ClosesAtMinute:         closes,
		PreorderMinLeadMinutes: int32(lead),
		MaxAdvanceDays:         advance,
		SlotMinutes:            slot,
		SlotCapacity:           capacity,
		InstantPrepMinutes:     prep,
		PaymentHoldMinutes:     hold,
	}, nil
}

func int16Of(field string, v int64) (int16, error) {
	if v < math.MinInt16 || v > math.MaxInt16 {
		return 0, apperrors.Errorf("%s %d out of range", field, v)
	}
	return int16(v), nil
}

var _ port.StoreSettingsRepository = (*StoreSettingsRepository)(nil)
