package order

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainstore "github.com/boms/backend/internal/domain/store"
)

// pickupCases is contracts/pickup-rules-cases.json, which the frontend's
// pickupProblem and pickupOrderType read too.
type pickupCases struct {
	Policy struct {
		Rules struct {
			OpensAt         string   `json:"opens_at"`
			ClosesAt        string   `json:"closes_at"`
			PreorderMinLead int      `json:"preorder_min_lead_minutes"`
			MaxAdvanceDays  int      `json:"max_advance_days"`
			SlotMinutes     int      `json:"slot_minutes"`
			InstantPrep     int      `json:"instant_prep_minutes"`
			ClosedDates     []string `json:"closed_dates"`
		} `json:"rules"`
		Now   string `json:"now"`
		Cases []struct {
			At          string  `json:"at"`
			Kitchen     bool    `json:"kitchen"`
			LeadMinutes int     `json:"lead_minutes"`
			SoldOutOn   *string `json:"sold_out_on"`
			Problem     *string `json:"problem"`
			Type        *string `json:"type"`
		} `json:"cases"`
	} `json:"policy"`
}

const wallClockLayout = "2006-01-02T15:04"

func loadPickupCases(t *testing.T) pickupCases {
	t.Helper()
	// A rooted file system: the read cannot leave contracts/.
	raw, err := fs.ReadFile(os.DirFS(filepath.Join("..", "..", "..", "..", "contracts")), "pickup-rules-cases.json")
	require.NoError(t, err)
	var cases pickupCases
	require.NoError(t, json.Unmarshal(raw, &cases))
	require.NotEmpty(t, cases.Policy.Cases)
	return cases
}

func wallClock(t *testing.T, s string) time.Time {
	t.Helper()
	if s == "" {
		return time.Time{}
	}
	at, err := time.ParseInLocation(wallClockLayout, s, domainstore.Location)
	require.NoError(t, err)
	return at
}

// problemCode names a pickup error the way the contract fixture does.
func problemCode(err error) *string {
	codes := map[error]string{
		ErrInvalidPickupAt:    "missing",
		ErrPickupTooSoon:      "too_soon",
		ErrPickupTooFar:       "too_far",
		ErrPickupClosedDay:    "closed_day",
		ErrPickupOutsideHours: "outside_hours",
		ErrPickupOffSlot:      "off_slot",
		ErrPickupSoldOut:      "sold_out",
	}
	for sentinel, code := range codes {
		if errors.Is(err, sentinel) {
			return &code
		}
	}
	return nil
}

func policyFromRules(t *testing.T, cases pickupCases) PickupPolicy {
	t.Helper()
	rules := cases.Policy.Rules
	opens, err := domainstore.ParseClock(rules.OpensAt)
	require.NoError(t, err)
	closes, err := domainstore.ParseClock(rules.ClosesAt)
	require.NoError(t, err)
	closed := make([]time.Time, 0, len(rules.ClosedDates))
	for _, day := range rules.ClosedDates {
		parsed, err := time.Parse(domainstore.DayLayout, day)
		require.NoError(t, err)
		closed = append(closed, parsed)
	}
	return PickupPolicy{
		Settings: domainstore.Settings{
			OpensAt:         opens,
			ClosesAt:        closes,
			PreorderMinLead: time.Duration(rules.PreorderMinLead) * time.Minute,
			MaxAdvanceDays:  rules.MaxAdvanceDays,
			SlotLength:      time.Duration(rules.SlotMinutes) * time.Minute,
			SlotCapacity:    1,
			InstantPrep:     time.Duration(rules.InstantPrep) * time.Minute,
			// Not a pickup rule: any hold an admin could save.
			PaymentHold: domainstore.MinPaymentHold,
		},
		ClosedDays: closed,
	}
}

func TestPickupPolicy_Contract(t *testing.T) {
	t.Parallel()
	cases := loadPickupCases(t)
	policy := policyFromRules(t, cases)
	require.NoError(t, policy.Settings.Validate(), "the fixture's rules are rules an admin could save")
	now := wallClock(t, cases.Policy.Now)

	for _, tc := range cases.Policy.Cases {
		items := Fulfillment{Kitchen: tc.Kitchen, Lead: time.Duration(tc.LeadMinutes) * time.Minute}
		if tc.SoldOutOn != nil {
			day, err := time.Parse(domainstore.DayLayout, *tc.SoldOutOn)
			require.NoError(t, err)
			items.SoldOutOn = &day
		}
		orderType, err := policy.Validate(wallClock(t, tc.At), now, items)
		if tc.Problem == nil {
			require.NoError(t, err, "at %q", tc.At)
			require.NotNil(t, tc.Type, "at %q: an accepted case names its type", tc.At)
			assert.Equal(t, *tc.Type, string(orderType), "at %q", tc.At)
			continue
		}
		got := problemCode(err)
		require.NotNil(t, got, "at %q: want %s, got %v", tc.At, *tc.Problem, err)
		assert.Equal(t, *tc.Problem, *got, "at %q", tc.At)
		assert.Empty(t, orderType, "at %q: a refused time has no type", tc.At)
	}
}

func TestPickupPolicy_ReadsTimesInBakeryTime(t *testing.T) {
	t.Parallel()
	cases := loadPickupCases(t)
	policy := policyFromRules(t, cases)
	now := wallClock(t, cases.Policy.Now)
	kitchen := Fulfillment{Kitchen: true}

	// 07:00 UTC is 14:00 at the bakery; 20:00 UTC on the 12th is 03:00 on the closed 13th.
	_, err := policy.Validate(time.Date(2026, 7, 10, 7, 0, 0, 0, time.UTC), now, kitchen)
	require.NoError(t, err)
	_, err = policy.Validate(time.Date(2026, 7, 12, 20, 0, 0, 0, time.UTC), now, kitchen)
	assert.ErrorIs(t, err, ErrPickupClosedDay)
}

func TestPickupPolicy_SlotsStartOnWholeMinutes(t *testing.T) {
	t.Parallel()
	cases := loadPickupCases(t)
	policy := policyFromRules(t, cases)
	now := wallClock(t, cases.Policy.Now)

	_, err := policy.Validate(wallClock(t, "2026-07-11T12:00").Add(30*time.Second), now, Fulfillment{Kitchen: true})
	assert.ErrorIs(t, err, ErrPickupOffSlot)
}

func TestPickupPolicy_HonoursEditedHours(t *testing.T) {
	t.Parallel()
	cases := loadPickupCases(t)
	policy := policyFromRules(t, cases)
	now := wallClock(t, cases.Policy.Now)
	policy.Settings.OpensAt = 9*time.Hour + 30*time.Minute
	policy.Settings.ClosesAt = 22 * time.Hour
	kitchen := Fulfillment{Kitchen: true}

	_, err := policy.Validate(wallClock(t, "2026-07-11T09:00"), now, kitchen)
	assert.ErrorIs(t, err, ErrPickupOutsideHours)
	_, err = policy.Validate(wallClock(t, "2026-07-11T21:30"), now, kitchen)
	require.NoError(t, err, "slots count from the new opening time")
	_, err = policy.Validate(wallClock(t, "2026-07-11T21:00").Add(-15*time.Minute), now, kitchen)
	assert.ErrorIs(t, err, ErrPickupOffSlot)
}
