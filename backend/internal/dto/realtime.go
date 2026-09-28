package dto

// RealtimeTicketResponse tells the browser how to open its push connection:
// connect to URL with the ticket as the `ticket` query parameter, at once. The
// ticket works once and only for seconds.
type RealtimeTicketResponse struct {
	Ticket string `json:"ticket"`
	URL    string `json:"url"`
}
