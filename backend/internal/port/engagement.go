package port

import "context"

// EngagementCounts is how many open customer accounts there are, how many of
// them used each engagement feature, and how many used none, one, or two or
// more of them.
type EngagementCounts struct {
	Customers     int64
	Reviews       int64
	Favorites     int64
	Wishlist      int64
	Messages      int64
	Promotions    int64
	UsedNone      int64
	UsedOne       int64
	UsedTwoOrMore int64
}

// EngagementRepository reads how customers use the engagement features.
type EngagementRepository interface {
	CountEngagement(ctx context.Context) (EngagementCounts, error)
}
