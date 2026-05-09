package dto

// ReadinessResponse is returned by GET /ready.
type ReadinessResponse struct {
	Status string         `json:"status"`
	Checks []CheckOutcome `json:"checks"`
}

// CheckOutcome describes one dependency probe.
type CheckOutcome struct {
	Name  string `json:"name"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}
