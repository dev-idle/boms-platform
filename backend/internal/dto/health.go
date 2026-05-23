package dto

// ReadinessResponse is returned by GET /ready (sanitized — no internal error strings).
type ReadinessResponse struct {
	Status string          `json:"status"`
	Checks ReadinessChecks `json:"checks"`
}

// ReadinessChecks reports dependency probe outcomes using stable public values.
type ReadinessChecks struct {
	DB    string `json:"db"`
	Redis string `json:"redis"`
}
