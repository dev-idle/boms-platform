package dto

import "time"

// ReportOrderIncidentRequest is staff reporting what went wrong with an order;
// the usecase checks the type and the note.
type ReportOrderIncidentRequest struct {
	Type string `json:"type" validate:"required"`
	Note string `json:"note" validate:"required"`
}

// OrderIncidentResponse is an incident recorded with an order. Source is
// "auto" when the system recorded it, "manual" when staff reported it; Note
// is what staff wrote reporting it.
type OrderIncidentResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	Note      *string   `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// ManagerOrderIncidentResponse is an incident in the manager's log, with its
// order and the staff member who recorded it, null when the system did. A
// bakery cancellation's note is the reason the customer was given.
type ManagerOrderIncidentResponse struct {
	OrderIncidentResponse
	OrderCode string     `json:"order_code"`
	PickupAt  *time.Time `json:"pickup_at"`
	ActorName *string    `json:"actor_name"`
}

// IncidentSummaryResponse is how many incidents of each type a week holds; a
// type with none is left out.
type IncidentSummaryResponse struct {
	Types []IncidentTypeCountResponse `json:"types"`
}

// IncidentTypeCountResponse is how many incidents of one type were recorded.
type IncidentTypeCountResponse struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}
