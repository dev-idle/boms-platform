package dto

import (
	"encoding/json"
	"time"
)

// DataExportResponse is everything the bakery holds about one person, as they
// download it from their account: the account and profile, the policies they
// accepted at sign-up, and every order they placed.
type DataExportResponse struct {
	ExportedAt      time.Time                    `json:"exported_at"`
	Account         MeResponse                   `json:"account"`
	TermsAcceptance *TermsAcceptanceResponse     `json:"terms_acceptance"`
	Sessions        []DataExportSessionResponse  `json:"sessions"`
	AccountActivity []AccountActivityResponse    `json:"account_activity"`
	Cart            []DataExportCartItemResponse `json:"cart"`
	Orders          []DataExportOrderResponse    `json:"orders"`
}

// DataExportSessionResponse is one place the person is signed in: since when,
// and from which network and browser — never the session's secrets.
type DataExportSessionResponse struct {
	SignedInAt time.Time `json:"signed_in_at"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
}

// AccountActivityByYou is who made a change the account holder made.
const AccountActivityByYou = "you"

// AccountActivityResponse is one recorded change to the account.
type AccountActivityResponse struct {
	Action string    `json:"action"`
	At     time.Time `json:"at"`
	// By is "you" for the account holder's own changes and the role of the
	// staff member otherwise.
	By     string          `json:"by"`
	Before json.RawMessage `json:"before"`
	After  json.RawMessage `json:"after"`
	// IP and UserAgent appear only on the account holder's own changes.
	IP        *string `json:"ip,omitempty"`
	UserAgent *string `json:"user_agent,omitempty"`
}

// DataExportCartItemResponse is a line waiting in the customer's cart.
type DataExportCartItemResponse struct {
	ID            string          `json:"id"`
	LineType      string          `json:"line_type"`
	ProductID     *string         `json:"product_id,omitempty"`
	ComboID       *string         `json:"combo_id,omitempty"`
	Quantity      int32           `json:"quantity"`
	Configuration json.RawMessage `json:"configuration,omitempty"`
	AddedAt       time.Time       `json:"added_at"`
}

// TermsAcceptanceResponse is the policy version someone accepted and when.
type TermsAcceptanceResponse struct {
	Version    string    `json:"version"`
	AcceptedAt time.Time `json:"accepted_at"`
}

// DataExportOrderResponse is one order as its customer sees it, with the
// policies accepted when it was placed.
type DataExportOrderResponse struct {
	ID                   string                       `json:"id"`
	Code                 string                       `json:"code"`
	Status               string                       `json:"status"`
	OrderType            string                       `json:"order_type"`
	SubtotalCents        int64                        `json:"subtotal_cents"`
	DiscountCents        int64                        `json:"discount_cents"`
	TotalCents           int64                        `json:"total_cents"`
	DiscountCodeSnapshot *string                      `json:"discount_code_snapshot,omitempty"`
	PickupAt             *time.Time                   `json:"pickup_at,omitempty"`
	TermsAcceptance      *TermsAcceptanceResponse     `json:"terms_acceptance"`
	Items                []OrderItemResponse          `json:"items"`
	Timeline             []OrderTimelineEntryResponse `json:"timeline"`
	CreatedAt            time.Time                    `json:"created_at"`
	UpdatedAt            time.Time                    `json:"updated_at"`
}
