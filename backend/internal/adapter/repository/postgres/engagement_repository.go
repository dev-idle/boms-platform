package postgres

import (
	"context"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	"github.com/boms/backend/internal/port"
)

// EngagementRepository reads how customers use the engagement features.
type EngagementRepository struct {
	queries *sqlcgen.Queries
}

func NewEngagementRepository(pool *Pool) *EngagementRepository {
	return &EngagementRepository{queries: pool.Queries()}
}

var _ port.EngagementRepository = (*EngagementRepository)(nil)

func (r *EngagementRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// CountEngagement implements port.EngagementRepository.
func (r *EngagementRepository) CountEngagement(ctx context.Context) (port.EngagementCounts, error) {
	row, err := r.q(ctx).CountEngagement(ctx)
	if err != nil {
		return port.EngagementCounts{}, mapRepoError(err, "count engagement")
	}
	return port.EngagementCounts{
		Customers:     row.Customers,
		Reviews:       row.Reviews,
		Favorites:     row.Favorites,
		Wishlist:      row.Wishlist,
		Messages:      row.Messages,
		Promotions:    row.Promotions,
		UsedNone:      row.UsedNone,
		UsedOne:       row.UsedOne,
		UsedTwoOrMore: row.UsedTwoOrMore,
	}, nil
}
