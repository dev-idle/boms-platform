package dto

import "time"

// TicketItemResponse is one product a ticket makes.
type TicketItemResponse struct {
	Name     string `json:"name"`
	Quantity int32  `json:"quantity"`
}

// TicketSummaryResponse is where one station's part of an order stands.
type TicketSummaryResponse struct {
	Station string `json:"station"`
	Status  string `json:"status"`
}

// OrderTicketResponse is a ticket in the counter's view of an order: what each
// station makes, so staff can move a ticket nobody has started.
type OrderTicketResponse struct {
	ID      string               `json:"id"`
	Station string               `json:"station"`
	Status  string               `json:"status"`
	Items   []TicketItemResponse `json:"items"`
}

// TicketCustomerResponse names the order for a station and nothing more: the
// stations never contact the customer, so no email or phone reaches them.
type TicketCustomerResponse struct {
	DisplayName *string `json:"display_name,omitempty"`
}

// StationTicketResponse is a ticket in a station's queue.
type StationTicketResponse struct {
	ID          string                 `json:"id"`
	OrderID     string                 `json:"order_id"`
	OrderCode   string                 `json:"order_code"`
	OrderStatus string                 `json:"order_status"`
	Station     string                 `json:"station"`
	Status      string                 `json:"status"`
	PickupAt    *time.Time             `json:"pickup_at,omitempty"`
	Customer    TicketCustomerResponse `json:"customer"`
	Items       []TicketItemResponse   `json:"items"`
	CreatedAt   time.Time              `json:"created_at"`
}

// StationTicketDetailResponse is a station's ticket with where the order's
// other tickets stand, so the station knows whether the order waits on it.
type StationTicketDetailResponse struct {
	StationTicketResponse
	OrderTickets []TicketSummaryResponse `json:"order_tickets"`
}

// TicketChangeResponse is a ticket after a move, with the status its order reached.
type TicketChangeResponse struct {
	ID          string `json:"id"`
	OrderID     string `json:"order_id"`
	Station     string `json:"station"`
	Status      string `json:"status"`
	OrderStatus string `json:"order_status"`
}

type PatchTicketStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=in_progress ready"`
}

type MoveTicketRequest struct {
	Station string `json:"station" validate:"required,oneof=kitchen counter"`
}
