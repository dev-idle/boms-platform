package dto

// PickupRulesResponse is what a checkout picker needs to offer valid pickup
// times: the opening hours (HH:MM, bakery time), the slot length, the notice
// for pre-orders and for instant orders, the booking window, and the closed
// days inside it.
type PickupRulesResponse struct {
	OpensAt                string                     `json:"opens_at"`
	ClosesAt               string                     `json:"closes_at"`
	SlotMinutes            int                        `json:"slot_minutes"`
	PreorderMinLeadMinutes int                        `json:"preorder_min_lead_minutes"`
	InstantPrepMinutes     int                        `json:"instant_prep_minutes"`
	MaxAdvanceDays         int                        `json:"max_advance_days"`
	ClosedDates            []PublicClosedDateResponse `json:"closed_dates"`
}

// PickupSlotsResponse is one day's pickup slots, each marked full when it takes
// no more orders. A closed day has none.
type PickupSlotsResponse struct {
	Date  string               `json:"date"`
	Slots []PickupSlotResponse `json:"slots"`
}

// PickupSlotResponse is a pickup slot by its start (RFC 3339).
type PickupSlotResponse struct {
	StartsAt string `json:"starts_at"`
	Full     bool   `json:"full"`
}

// PublicClosedDateResponse is a closed day as customers see it.
type PublicClosedDateResponse struct {
	Date   string `json:"date"`
	Reason string `json:"reason"`
}

// StoreSettingsResponse is the settings an admin edits.
type StoreSettingsResponse struct {
	OpensAt                string `json:"opens_at"`
	ClosesAt               string `json:"closes_at"`
	PreorderMinLeadMinutes int    `json:"preorder_min_lead_minutes"`
	MaxAdvanceDays         int    `json:"max_advance_days"`
	SlotMinutes            int    `json:"slot_minutes"`
	SlotCapacity           int    `json:"slot_capacity"`
	InstantPrepMinutes     int    `json:"instant_prep_minutes"`
	PaymentHoldMinutes     int    `json:"payment_hold_minutes"`
	UpdatedAt              string `json:"updated_at"`
}

// PatchStoreSettingsRequest changes the settings it names; absent fields stay.
type PatchStoreSettingsRequest struct {
	OpensAt                *string `json:"opens_at"`
	ClosesAt               *string `json:"closes_at"`
	PreorderMinLeadMinutes *int    `json:"preorder_min_lead_minutes"`
	MaxAdvanceDays         *int    `json:"max_advance_days"`
	SlotMinutes            *int    `json:"slot_minutes"`
	SlotCapacity           *int    `json:"slot_capacity"`
	InstantPrepMinutes     *int    `json:"instant_prep_minutes"`
	PaymentHoldMinutes     *int    `json:"payment_hold_minutes"`
}

// ClosedDateResponse is a closed day as admins manage it.
type ClosedDateResponse struct {
	ID        string `json:"id"`
	Date      string `json:"date"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at"`
}

// CreateClosedDateRequest closes the bakery on Date (YYYY-MM-DD, bakery time).
type CreateClosedDateRequest struct {
	Date   string `json:"date" validate:"required"`
	Reason string `json:"reason" validate:"required,max=200"`
}
