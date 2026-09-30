package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// AdminStoreSettingsUsecase lets admins change the pickup rules without a
// deploy. Every change is audited and announced to open pages as it commits.
type AdminStoreSettingsUsecase struct {
	store  port.StoreSettingsRepository
	tx     port.TxManager
	events port.EventOutbox
	audit  *auditlogger.Service
	log    *zap.Logger
}

func NewAdminStoreSettingsUsecase(
	store port.StoreSettingsRepository,
	tx port.TxManager,
	events port.EventOutbox,
	audit *auditlogger.Service,
	log *zap.Logger,
) *AdminStoreSettingsUsecase {
	return &AdminStoreSettingsUsecase{store: store, tx: tx, events: events, audit: audit, log: log}
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
