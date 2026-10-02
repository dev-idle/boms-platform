package usecase

import (
	"context"

	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
)

// ManagerEngagementUsecase is a manager reading how customers use the
// engagement features: reviews, favorites, the wishlist, messages and
// promotion emails.
type ManagerEngagementUsecase struct {
	engagement port.EngagementRepository
}

func NewManagerEngagementUsecase(engagement port.EngagementRepository) *ManagerEngagementUsecase {
	return &ManagerEngagementUsecase{engagement: engagement}
}

// Report counts the open customer accounts, how many used none, one, or two
// or more of the features, and how many used each.
func (u *ManagerEngagementUsecase) Report(ctx context.Context) (*dto.EngagementReportResponse, error) {
	counts, err := u.engagement.CountEngagement(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.EngagementReportResponse{
		Customers:     counts.Customers,
		UsedNone:      counts.UsedNone,
		UsedOne:       counts.UsedOne,
		UsedTwoOrMore: counts.UsedTwoOrMore,
		Features: []dto.EngagementFeatureResponse{
			{Feature: "reviews", Customers: counts.Reviews},
			{Feature: "favorites", Customers: counts.Favorites},
			{Feature: "wishlist", Customers: counts.Wishlist},
			{Feature: "messages", Customers: counts.Messages},
			{Feature: "promotions", Customers: counts.Promotions},
		},
	}, nil
}
