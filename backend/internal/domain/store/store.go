// Package store holds the bakery's operating settings: when it hands out
// pickups and which days it is closed. Admins edit them; checkout enforces them.
package store

import (
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Location is the bakery's time zone. It is fixed rather than a setting: the
// bakery does not move, and every stored hour is read in this zone.
// Asia/Ho_Chi_Minh keeps one offset all year.
var Location = time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)

// DayLayout is how a calendar day is written in the API.
const DayLayout = "2006-01-02"

// MaxAdvanceDaysLimit is the longest booking window an admin may set, so the
// closed days a checkout can meet are known before the settings are read.
const MaxAdvanceDaysLimit = 90

// MaxPreorderMinLead is the longest pre-order notice an admin may require.
const MaxPreorderMinLead = 7 * 24 * time.Hour

// slotLengths are the pickup slot lengths an admin may choose: each divides an
// hour, so slots start at times people say.
var slotLengths = []time.Duration{10 * time.Minute, 15 * time.Minute, 20 * time.Minute, 30 * time.Minute, time.Hour}

const (
	// MaxSlotCapacity is the most orders one pickup slot may take.
	MaxSlotCapacity = 200
	// MaxInstantPrep is the longest the counter may take to pack an instant order.
	MaxInstantPrep = 4 * time.Hour
	// MinPaymentHold and MaxPaymentHold bound how long an order placed online
	// may hold its slot unpaid: long enough to pay, short enough not to keep
	// a slot from other customers.
	MinPaymentHold = 5 * time.Minute
	MaxPaymentHold = 2 * time.Hour
)

const (
	minAdvanceDays = 1
	// maxClosedDateAhead bounds how far ahead a closure may be planned, which
	// also bounds any list of closed dates.
	maxClosedDateAhead = 365
	maxReasonLength    = 200
)

// Settings are the pickup rules in force. Hours are offsets from local
// midnight within one day: pickups start on SlotLength marks from OpensAt, up
// to, not including, ClosesAt. SlotCapacity orders may share a slot.
// InstantPrep is how long the counter needs to pack a same-day order of
// ready-made items; every other order waits PreorderMinLead. An order placed
// online holds its slot and discount unpaid for PaymentHold.
type Settings struct {
	OpensAt         time.Duration
	ClosesAt        time.Duration
	PreorderMinLead time.Duration
	MaxAdvanceDays  int
	SlotLength      time.Duration
	SlotCapacity    int
	InstantPrep     time.Duration
	PaymentHold     time.Duration
	UpdatedAt       time.Time
}

// Validate reports the first rule the settings break.
func (s Settings) Validate() error {
	if !wholeMinutes(s.OpensAt) || !wholeMinutes(s.ClosesAt) ||
		s.OpensAt < 0 || s.ClosesAt >= 24*time.Hour || s.OpensAt >= s.ClosesAt {
		return ErrInvalidHours
	}
	if s.MaxAdvanceDays < minAdvanceDays || s.MaxAdvanceDays > MaxAdvanceDaysLimit {
		return ErrInvalidAdvanceDays
	}
	// A lead time as long as the booking window would leave no pickup time.
	if !wholeMinutes(s.PreorderMinLead) || s.PreorderMinLead < 0 ||
		s.PreorderMinLead > MaxPreorderMinLead ||
		s.PreorderMinLead >= time.Duration(s.MaxAdvanceDays)*24*time.Hour {
		return ErrInvalidLeadTime
	}
	if !slices.Contains(slotLengths, s.SlotLength) || s.ClosesAt-s.OpensAt < s.SlotLength {
		return ErrInvalidSlotLength
	}
	if s.SlotCapacity < 1 || s.SlotCapacity > MaxSlotCapacity {
		return ErrInvalidSlotCapacity
	}
	if !wholeMinutes(s.InstantPrep) || s.InstantPrep < 0 || s.InstantPrep > MaxInstantPrep {
		return ErrInvalidInstantPrep
	}
	if !wholeMinutes(s.PaymentHold) || s.PaymentHold < MinPaymentHold || s.PaymentHold > MaxPaymentHold {
		return ErrInvalidPaymentHold
	}
	return nil
}

// MissedBefore is the instant before which a pickup not collected by now is
// missed: the start of the day after the last closing time now has passed.
func (s Settings) MissedBefore(now time.Time) time.Time {
	y, m, d := now.In(Location).Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, Location)
	if now.Before(today.Add(s.ClosesAt)) {
		return today
	}
	return today.AddDate(0, 0, 1)
}

// LastPickupDay is the last bakery day the booking window reaches from now.
func (s Settings) LastPickupDay(now time.Time) time.Time {
	return DayOf(now.Add(time.Duration(s.MaxAdvanceDays) * 24 * time.Hour))
}

// SlotStartOf is the start of the slot that at falls in, and whether it falls
// in one at all (inside the opening hours). An order booked before the hours
// or the slot length changed still counts against the slot now covering it.
func (s Settings) SlotStartOf(at time.Time) (time.Time, bool) {
	local := at.In(Location)
	midnight := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, Location)
	since := local.Sub(midnight)
	if since < s.OpensAt || since >= s.ClosesAt {
		return time.Time{}, false
	}
	return midnight.Add(s.OpensAt + (since-s.OpensAt)/s.SlotLength*s.SlotLength), true
}

// SlotStarts are the pickup times of a bakery day (as DayOf returns it), in
// order: OpensAt and every SlotLength after it that starts before ClosesAt.
func (s Settings) SlotStarts(day time.Time) []time.Time {
	midnight := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, Location)
	starts := make([]time.Time, 0, int((s.ClosesAt-s.OpensAt)/s.SlotLength)+1)
	for offset := s.OpensAt; offset < s.ClosesAt; offset += s.SlotLength {
		starts = append(starts, midnight.Add(offset))
	}
	return starts
}

func wholeMinutes(d time.Duration) bool {
	return d%time.Minute == 0
}

// DayOf is the bakery-local calendar day of t, held at midnight UTC so it
// compares and stores as the plain date it is.
func DayOf(t time.Time) time.Time {
	y, m, d := t.In(Location).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// ClosedDate is a day the bakery takes no pickups, with the reason customers see.
type ClosedDate struct {
	ID        uuid.UUID
	Day       time.Time
	Reason    string
	CreatedAt time.Time
}

// ClosedDateWindow is the span of days a closure may be planned for, counted
// from the bakery-local day of now: today through a year ahead.
func ClosedDateWindow(now time.Time) (from, to time.Time) {
	from = DayOf(now)
	return from, from.AddDate(0, 0, maxClosedDateAhead)
}

// NewClosedDate checks a closure an admin is adding and returns its trimmed reason.
func NewClosedDate(day time.Time, reason string, now time.Time) (string, error) {
	from, to := ClosedDateWindow(now)
	if day.Before(from) || day.After(to) {
		return "", ErrClosedDateOutOfRange
	}
	reason = strings.TrimSpace(reason)
	if reason == "" || utf8.RuneCountInString(reason) > maxReasonLength || !PlainText(reason) {
		return "", ErrInvalidClosedDateReason
	}
	return reason, nil
}

// PlainText rejects control and invisible formatting characters: text a
// customer reads must read as it was saved, and a direction override or a
// zero-width character would make it read differently.
func PlainText(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return false
		}
	}
	return true
}

// beyondPlainText is what free text written on a phone may hold that plain
// text does not: line breaks, and the zero-width joiner emoji such as a cook
// or a shrug are built from.
var beyondPlainText = strings.NewReplacer("\n", "", string(rune(0x200d)), "")

// PlainLines is PlainText for text that may break across lines, as a message
// or a review is written.
func PlainLines(s string) bool {
	return PlainText(beyondPlainText.Replace(s))
}

// TrimLines trims text that may break across lines, with every line break
// stored as "\n".
func TrimLines(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
}
