package store

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validSettings() Settings {
	return Settings{
		OpensAt: 8 * time.Hour, ClosesAt: 18 * time.Hour, PreorderMinLead: 2 * time.Hour, MaxAdvanceDays: 14,
		SlotLength: 30 * time.Minute, SlotCapacity: 10, InstantPrep: 20 * time.Minute,
	}
}

func TestSettingsValidate(t *testing.T) {
	t.Parallel()
	require.NoError(t, validSettings().Validate())

	cases := map[string]struct {
		change func(*Settings)
		want   error
	}{
		"opens_at_closing":           {func(s *Settings) { s.OpensAt = s.ClosesAt }, ErrInvalidHours},
		"opens_after_closing":        {func(s *Settings) { s.OpensAt = 19 * time.Hour }, ErrInvalidHours},
		"closes_at_midnight":         {func(s *Settings) { s.ClosesAt = 24 * time.Hour }, ErrInvalidHours},
		"opens_before_midnight":      {func(s *Settings) { s.OpensAt = -time.Minute }, ErrInvalidHours},
		"hours_with_seconds":         {func(s *Settings) { s.OpensAt = 8*time.Hour + time.Second }, ErrInvalidHours},
		"no_booking_window":          {func(s *Settings) { s.MaxAdvanceDays = 0 }, ErrInvalidAdvanceDays},
		"booking_window_too_long":    {func(s *Settings) { s.MaxAdvanceDays = 91 }, ErrInvalidAdvanceDays},
		"negative_lead_time":         {func(s *Settings) { s.PreorderMinLead = -time.Minute }, ErrInvalidLeadTime},
		"lead_time_over_a_week":      {func(s *Settings) { s.PreorderMinLead = 7*24*time.Hour + time.Minute }, ErrInvalidLeadTime},
		"lead_time_fills_the_window": {func(s *Settings) { s.MaxAdvanceDays = 1; s.PreorderMinLead = 24 * time.Hour }, ErrInvalidLeadTime},
		"slot_not_dividing_an_hour":  {func(s *Settings) { s.SlotLength = 45 * time.Minute }, ErrInvalidSlotLength},
		"no_slot_length":             {func(s *Settings) { s.SlotLength = 0 }, ErrInvalidSlotLength},
		"slot_longer_than_the_hours": {func(s *Settings) { s.OpensAt = 17*time.Hour + 30*time.Minute; s.SlotLength = time.Hour }, ErrInvalidSlotLength},
		"slot_takes_no_orders":       {func(s *Settings) { s.SlotCapacity = 0 }, ErrInvalidSlotCapacity},
		"slot_takes_too_many":        {func(s *Settings) { s.SlotCapacity = MaxSlotCapacity + 1 }, ErrInvalidSlotCapacity},
		"negative_instant_prep":      {func(s *Settings) { s.InstantPrep = -time.Minute }, ErrInvalidInstantPrep},
		"instant_prep_over_4_hours":  {func(s *Settings) { s.InstantPrep = MaxInstantPrep + time.Minute }, ErrInvalidInstantPrep},
		"instant_prep_with_seconds":  {func(s *Settings) { s.InstantPrep = time.Minute + time.Second }, ErrInvalidInstantPrep},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := validSettings()
			tc.change(&s)
			assert.ErrorIs(t, s.Validate(), tc.want)
		})
	}

	t.Run("allows_open_all_day_and_no_lead_time", func(t *testing.T) {
		t.Parallel()
		s := Settings{
			OpensAt: 0, ClosesAt: 23*time.Hour + 59*time.Minute, PreorderMinLead: 0, MaxAdvanceDays: 1,
			SlotLength: 10 * time.Minute, SlotCapacity: 1, InstantPrep: 0,
		}
		require.NoError(t, s.Validate())
	})
}

func TestSettingsSlotStartOf(t *testing.T) {
	t.Parallel()
	s := validSettings()
	s.OpensAt = 8*time.Hour + 15*time.Minute
	at := func(clock string) time.Time {
		parsed, err := time.ParseInLocation("2006-01-02T15:04", "2026-07-10T"+clock, Location)
		require.NoError(t, err)
		return parsed
	}

	for clock, want := range map[string]string{"08:15": "08:15", "08:30": "08:15", "08:44": "08:15", "08:45": "08:45", "17:59": "17:45"} {
		start, ok := s.SlotStartOf(at(clock))
		require.True(t, ok, clock)
		assert.Equal(t, want, start.In(Location).Format("15:04"), "an order at %s", clock)
	}
	for _, clock := range []string{"08:14", "18:00", "23:00"} {
		_, ok := s.SlotStartOf(at(clock))
		assert.False(t, ok, "%s is outside the hours", clock)
	}
}

func TestSettingsLastPickupDay(t *testing.T) {
	t.Parallel()
	s := validSettings()
	// 23:30 at the bakery on the 10th: fourteen days on is the 24th.
	now := time.Date(2026, 7, 10, 23, 30, 0, 0, Location)
	assert.Equal(t, time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC), s.LastPickupDay(now))
}

func TestSettingsSlotStarts(t *testing.T) {
	t.Parallel()
	s := validSettings()
	s.OpensAt = 8*time.Hour + 15*time.Minute
	s.ClosesAt = 10 * time.Hour

	starts := s.SlotStarts(time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC))

	got := make([]string, 0, len(starts))
	for _, start := range starts {
		got = append(got, start.In(Location).Format("2006-01-02T15:04"))
	}
	assert.Equal(t, []string{"2026-07-10T08:15", "2026-07-10T08:45", "2026-07-10T09:15", "2026-07-10T09:45"}, got,
		"slots count from opening; the last one starts before closing")
}

func TestDayOf(t *testing.T) {
	t.Parallel()
	// 18:30 UTC on the 12th is 01:30 on the 13th at the bakery.
	got := DayOf(time.Date(2026, 7, 12, 18, 30, 0, 0, time.UTC))
	assert.Equal(t, time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC), got)
}

func TestNewClosedDate(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 10, 23, 0, 0, 0, Location)
	today := DayOf(now)

	t.Run("accepts_today_through_a_year_ahead", func(t *testing.T) {
		t.Parallel()
		for _, day := range []time.Time{today, today.AddDate(0, 0, 365)} {
			reason, err := NewClosedDate(day, "  Lunar New Year  ", now)
			require.NoError(t, err)
			assert.Equal(t, "Lunar New Year", reason)
		}
	})

	t.Run("refuses_the_past_and_beyond_a_year", func(t *testing.T) {
		t.Parallel()
		for _, day := range []time.Time{today.AddDate(0, 0, -1), today.AddDate(0, 0, 366)} {
			_, err := NewClosedDate(day, "Holiday", now)
			assert.ErrorIs(t, err, ErrClosedDateOutOfRange)
		}
	})

	t.Run("needs_a_short_plain_reason", func(t *testing.T) {
		t.Parallel()
		for _, reason := range []string{
			"", "   ", strings.Repeat("a", 201),
			"\x00", "Closed\nearly", "Tet \u202eyadiloh", "Zero\u200bwidth",
		} {
			_, err := NewClosedDate(today, reason, now)
			assert.ErrorIs(t, err, ErrInvalidClosedDateReason, "%q", reason)
		}
		_, err := NewClosedDate(today, strings.Repeat("é", 200), now)
		require.NoError(t, err, "the limit counts characters, not bytes")
	})
}
