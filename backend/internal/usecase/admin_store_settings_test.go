package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

func ptr[T any](v T) *T { return &v }

// requireValidationField asserts err is a validation error naming field.
func requireValidationField(t *testing.T, err error, field string) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.ErrValidation.Code, appErr.Code)
	assert.Contains(t, appErr.Details, field)
}

func TestAdminStoreSettingsUsecase_PatchSettings(t *testing.T) {
	t.Parallel()
	admin := uuid.New()

	setup := func() (*usecase.AdminStoreSettingsUsecase, *memoryStore, *recordingOutbox) {
		store, outbox := newMemoryStore(), &recordingOutbox{}
		return usecase.NewAdminStoreSettingsUsecase(store, passthroughTxManager{}, outbox, nil, nil), store, outbox
	}

	t.Run("changes_only_the_named_settings_and_announces_it", func(t *testing.T) {
		t.Parallel()
		uc, store, outbox := setup()

		out, err := uc.PatchSettings(context.Background(), admin, domainuser.RoleAdmin, dto.PatchStoreSettingsRequest{
			ClosesAt: ptr("21:30"), MaxAdvanceDays: ptr(30),
		})

		require.NoError(t, err)
		assert.Equal(t, "08:00", out.OpensAt)
		assert.Equal(t, "21:30", out.ClosesAt)
		assert.Equal(t, 120, out.PreorderMinLeadMinutes)
		assert.Equal(t, 30, out.MaxAdvanceDays)
		assert.Equal(t, 30, out.SlotMinutes, "untouched settings keep their values")
		assert.Equal(t, 21*time.Hour+30*time.Minute, store.settings.ClosesAt)
		require.Len(t, outbox.events, 1)
		assert.Equal(t, domainstore.TopicSettingsUpdated, outbox.events[0].Topic)
		assert.True(t, outbox.events[0].Audience.Public)
	})

	cases := map[string]struct {
		req  dto.PatchStoreSettingsRequest
		want error
	}{
		"closing_before_opening":     {dto.PatchStoreSettingsRequest{ClosesAt: ptr("07:00")}, domainstore.ErrInvalidHours},
		"booking_window_too_long":    {dto.PatchStoreSettingsRequest{MaxAdvanceDays: ptr(91)}, domainstore.ErrInvalidAdvanceDays},
		"lead_time_too_long":         {dto.PatchStoreSettingsRequest{PreorderMinLeadMinutes: ptr(7*24*60 + 1)}, domainstore.ErrInvalidLeadTime},
		"lead_time_that_would_wrap":  {dto.PatchStoreSettingsRequest{PreorderMinLeadMinutes: ptr(1 << 62)}, domainstore.ErrInvalidLeadTime},
		"lead_time_fills_the_window": {dto.PatchStoreSettingsRequest{MaxAdvanceDays: ptr(1), PreorderMinLeadMinutes: ptr(24 * 60)}, domainstore.ErrInvalidLeadTime},
		"negative_lead_time":         {dto.PatchStoreSettingsRequest{PreorderMinLeadMinutes: ptr(-1)}, domainstore.ErrInvalidLeadTime},
		"slot_of_45_minutes":         {dto.PatchStoreSettingsRequest{SlotMinutes: ptr(45)}, domainstore.ErrInvalidSlotLength},
		"slot_that_would_wrap":       {dto.PatchStoreSettingsRequest{SlotMinutes: ptr(1 << 62)}, domainstore.ErrInvalidSlotLength},
		"slot_without_room":          {dto.PatchStoreSettingsRequest{SlotCapacity: ptr(0)}, domainstore.ErrInvalidSlotCapacity},
		"instant_prep_too_long":      {dto.PatchStoreSettingsRequest{InstantPrepMinutes: ptr(241)}, domainstore.ErrInvalidInstantPrep},
		"instant_prep_that_wraps":    {dto.PatchStoreSettingsRequest{InstantPrepMinutes: ptr(1 << 62)}, domainstore.ErrInvalidInstantPrep},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			uc, store, outbox := setup()
			before := store.settings

			_, err := uc.PatchSettings(context.Background(), admin, domainuser.RoleAdmin, tc.req)

			require.ErrorIs(t, err, tc.want)
			assert.Equal(t, before, store.settings, "a refused change leaves the settings alone")
			assert.Empty(t, outbox.events)
		})
	}
}

func TestAdminStoreSettingsUsecase_PatchSettings_Slots(t *testing.T) {
	t.Parallel()
	store := newMemoryStore()
	uc := usecase.NewAdminStoreSettingsUsecase(store, passthroughTxManager{}, &recordingOutbox{}, nil, nil)

	out, err := uc.PatchSettings(context.Background(), uuid.New(), domainuser.RoleAdmin, dto.PatchStoreSettingsRequest{
		SlotMinutes: ptr(15), SlotCapacity: ptr(4), InstantPrepMinutes: ptr(45),
	})

	require.NoError(t, err)
	assert.Equal(t, 15, out.SlotMinutes)
	assert.Equal(t, 4, out.SlotCapacity)
	assert.Equal(t, 45, out.InstantPrepMinutes)
	assert.Equal(t, 15*time.Minute, store.settings.SlotLength)
	assert.Equal(t, 4, store.settings.SlotCapacity)
	assert.Equal(t, 45*time.Minute, store.settings.InstantPrep)
}

func TestAdminStoreSettingsUsecase_PatchSettings_Unreadable(t *testing.T) {
	t.Parallel()
	for field, req := range map[string]dto.PatchStoreSettingsRequest{
		"body":      {},
		"opens_at":  {OpensAt: ptr("8am")},
		"closes_at": {ClosesAt: ptr("25:00")},
	} {
		uc := usecase.NewAdminStoreSettingsUsecase(newMemoryStore(), passthroughTxManager{}, &recordingOutbox{}, nil, nil)
		_, err := uc.PatchSettings(context.Background(), uuid.New(), domainuser.RoleAdmin, req)
		requireValidationField(t, err, field)
	}
}

func TestAdminStoreSettingsUsecase_ClosedDates(t *testing.T) {
	t.Parallel()
	admin := uuid.New()
	tomorrow := domainstore.DayOf(time.Now()).AddDate(0, 0, 1).Format(domainstore.DayLayout)

	t.Run("closes_a_day_once_and_reopens_it", func(t *testing.T) {
		t.Parallel()
		store, outbox := newMemoryStore(), &recordingOutbox{}
		uc := usecase.NewAdminStoreSettingsUsecase(store, passthroughTxManager{}, outbox, nil, nil)
		ctx := context.Background()

		added, err := uc.AddClosedDate(ctx, admin, domainuser.RoleAdmin, dto.CreateClosedDateRequest{Date: tomorrow, Reason: " Staff training "})
		require.NoError(t, err)
		assert.Equal(t, tomorrow, added.Date)
		assert.Equal(t, "Staff training", added.Reason)

		_, err = uc.AddClosedDate(ctx, admin, domainuser.RoleAdmin, dto.CreateClosedDateRequest{Date: tomorrow, Reason: "Again"})
		require.ErrorIs(t, err, domainstore.ErrClosedDateExists)

		listed, err := uc.ListClosedDates(ctx)
		require.NoError(t, err)
		require.Len(t, listed, 1)

		require.NoError(t, uc.RemoveClosedDate(ctx, admin, domainuser.RoleAdmin, uuid.MustParse(added.ID)))
		require.ErrorIs(t, uc.RemoveClosedDate(ctx, admin, domainuser.RoleAdmin, uuid.MustParse(added.ID)), domainstore.ErrClosedDateNotFound)
		assert.Len(t, outbox.events, 2, "adding and removing each announce the change")
	})

	t.Run("refuses_an_unreadable_or_past_day", func(t *testing.T) {
		t.Parallel()
		uc := usecase.NewAdminStoreSettingsUsecase(newMemoryStore(), passthroughTxManager{}, &recordingOutbox{}, nil, nil)
		ctx := context.Background()

		_, err := uc.AddClosedDate(ctx, admin, domainuser.RoleAdmin, dto.CreateClosedDateRequest{Date: "10/02/2031", Reason: "Holiday"})
		requireValidationField(t, err, "date")
		_, err = uc.AddClosedDate(ctx, admin, domainuser.RoleAdmin, dto.CreateClosedDateRequest{Date: "2020-01-01", Reason: "Holiday"})
		require.ErrorIs(t, err, domainstore.ErrClosedDateOutOfRange)
	})
}
