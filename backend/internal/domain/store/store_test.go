package store

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validSettings() Settings {
	return Settings{OpensAt: 8 * time.Hour, ClosesAt: 18 * time.Hour, PreorderMinLead: 2 * time.Hour, MaxAdvanceDays: 14}
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
		s := Settings{OpensAt: 0, ClosesAt: 23*time.Hour + 59*time.Minute, PreorderMinLead: 0, MaxAdvanceDays: 1}
		require.NoError(t, s.Validate())
	})
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
