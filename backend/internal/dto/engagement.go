package dto

// EngagementReportResponse is how open customer accounts use the engagement
// features: how many there are, how many used none, one, or two or more, and
// how many used each.
type EngagementReportResponse struct {
	Customers     int64                       `json:"customers"`
	UsedNone      int64                       `json:"used_none"`
	UsedOne       int64                       `json:"used_one"`
	UsedTwoOrMore int64                       `json:"used_two_or_more"`
	Features      []EngagementFeatureResponse `json:"features"`
}

// EngagementFeatureResponse is how many customers used one feature.
type EngagementFeatureResponse struct {
	Feature   string `json:"feature"`
	Customers int64  `json:"customers"`
}
