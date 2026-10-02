package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainreview "github.com/boms/backend/internal/domain/review"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// ReviewRepository keeps customers' reviews of the products they picked up.
type ReviewRepository struct {
	queries *sqlcgen.Queries
}

func NewReviewRepository(pool *Pool) *ReviewRepository {
	return &ReviewRepository{queries: pool.Queries()}
}

func (r *ReviewRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// Create implements port.ReviewRepository.
func (r *ReviewRepository) Create(ctx context.Context, params port.CreateReviewParams) (*domainreview.Review, error) {
	row, err := r.q(ctx).CreateReview(ctx, sqlcgen.CreateReviewParams{
		OrderID:   params.OrderID,
		ProductID: params.ProductID,
		UserID:    params.UserID,
		Rating:    params.Rating,
		Comment:   params.Comment,
	})
	if err != nil {
		return nil, mapRepoError(err, "create review")
	}
	review := mapReview(sqlcgen.ListReviewsByOrderRow(row))
	return &review, nil
}

// ListByOrder implements port.ReviewRepository.
func (r *ReviewRepository) ListByOrder(ctx context.Context, userID, orderID uuid.UUID) ([]domainreview.Review, error) {
	rows, err := r.q(ctx).ListReviewsByOrder(ctx, sqlcgen.ListReviewsByOrderParams{OrderID: orderID, UserID: userID})
	if err != nil {
		return nil, mapRepoError(err, "list order reviews")
	}
	out := make([]domainreview.Review, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapReview(row))
	}
	return out, nil
}

// ListPublished implements port.ReviewRepository.
func (r *ReviewRepository) ListPublished(
	ctx context.Context,
	productID uuid.UUID,
	beforeID *uuid.UUID,
	limit int32,
) ([]port.PublishedReview, error) {
	rows, err := r.q(ctx).ListPublishedReviews(ctx, sqlcgen.ListPublishedReviewsParams{
		ProductID: productID,
		BeforeID:  beforeID,
		Limit:     limit,
	})
	if err != nil {
		return nil, mapRepoError(err, "list published reviews")
	}
	out := make([]port.PublishedReview, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.PublishedReview(row))
	}
	return out, nil
}

// PublishedRating implements port.ReviewRepository.
func (r *ReviewRepository) PublishedRating(ctx context.Context, productID uuid.UUID) (port.RatingTotal, error) {
	row, err := r.q(ctx).CountPublishedReviews(ctx, productID)
	if err != nil {
		return port.RatingTotal{}, mapRepoError(err, "count published reviews")
	}
	return port.RatingTotal{Count: row.ReviewCount, Total: row.RatingTotal}, nil
}

// ManagerList implements port.ReviewRepository.
func (r *ReviewRepository) ManagerList(ctx context.Context, params port.ManagerListReviewsParams) ([]port.ManagerReview, error) {
	rows, err := r.q(ctx).ManagerListReviews(ctx, sqlcgen.ManagerListReviewsParams{
		Status: optionalReviewStatus(params.Status),
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, mapRepoError(err, "list reviews")
	}
	out := make([]port.ManagerReview, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapManagerReview(row))
	}
	return out, nil
}

// ManagerListCount implements port.ReviewRepository.
func (r *ReviewRepository) ManagerListCount(ctx context.Context, status *domainreview.Status) (int64, error) {
	n, err := r.q(ctx).ManagerListReviewsCount(ctx, optionalReviewStatus(status))
	if err != nil {
		return 0, mapRepoError(err, "count reviews")
	}
	return n, nil
}

// GetForUpdate implements port.ReviewRepository.
func (r *ReviewRepository) GetForUpdate(ctx context.Context, id uuid.UUID) (port.LockedReview, error) {
	if txFromContext(ctx) == nil {
		return port.LockedReview{}, apperrors.Errorf("lock review: requires a transaction")
	}
	row, err := r.q(ctx).GetReviewForUpdate(ctx, id)
	if err != nil {
		return port.LockedReview{}, mapRepoError(err, "lock review")
	}
	return port.LockedReview{Status: domainreview.Status(row.Status), AuthorID: row.UserID}, nil
}

// Moderate implements port.ReviewRepository.
func (r *ReviewRepository) Moderate(
	ctx context.Context,
	id, moderatorID uuid.UUID,
	status domainreview.Status,
) (*port.ManagerReview, error) {
	row, err := r.q(ctx).ModerateReview(ctx, sqlcgen.ModerateReviewParams{
		ID:          id,
		Status:      sqlcgen.ReviewStatus(status),
		ModeratedBy: &moderatorID,
	})
	if err != nil {
		return nil, mapRepoError(err, "moderate review")
	}
	review := mapManagerReview(sqlcgen.ManagerListReviewsRow(row))
	return &review, nil
}

// RatingCounts implements port.ReviewRepository.
func (r *ReviewRepository) RatingCounts(ctx context.Context) (port.RatingCounts, error) {
	row, err := r.q(ctx).ReviewRatingCounts(ctx)
	if err != nil {
		return port.RatingCounts{}, mapRepoError(err, "count review ratings")
	}
	return port.RatingCounts{
		Count:   row.ReviewCount,
		Pending: row.PendingCount,
		Total:   row.RatingTotal,
		Stars:   [domainreview.MaxRating]int64{row.OneStar, row.TwoStars, row.ThreeStars, row.FourStars, row.FiveStars},
	}, nil
}

// RatingsByProduct implements port.ReviewRepository.
func (r *ReviewRepository) RatingsByProduct(ctx context.Context, limit int32) ([]port.ProductRating, error) {
	rows, err := r.q(ctx).ReviewRatingsByProduct(ctx, limit)
	if err != nil {
		return nil, mapRepoError(err, "list product ratings")
	}
	out := make([]port.ProductRating, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.ProductRating{
			ProductID:   row.ID,
			ProductName: row.Name,
			Count:       row.ReviewCount,
			Total:       row.RatingTotal,
		})
	}
	return out, nil
}

// ListForExport implements port.ReviewRepository.
func (r *ReviewRepository) ListForExport(ctx context.Context, userID uuid.UUID) ([]port.ReviewEntry, error) {
	rows, err := r.q(ctx).ListReviewsByUser(ctx, userID)
	if err != nil {
		return nil, mapRepoError(err, "list reviews for export")
	}
	out := make([]port.ReviewEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.ReviewEntry{
			OrderID:     row.OrderID,
			OrderCode:   row.OrderCode,
			ProductID:   row.ProductID,
			ProductName: row.ProductName,
			Rating:      row.Rating,
			Comment:     row.Comment,
			Status:      domainreview.Status(row.Status),
			CreatedAt:   row.CreatedAt,
		})
	}
	return out, nil
}

// EraseForCustomer implements port.ReviewRepository.
func (r *ReviewRepository) EraseForCustomer(ctx context.Context, userID uuid.UUID) error {
	if err := r.q(ctx).EraseUserReviews(ctx, userID); err != nil {
		return mapRepoError(err, "erase reviews")
	}
	return nil
}

func optionalReviewStatus(status *domainreview.Status) *sqlcgen.ReviewStatus {
	if status == nil {
		return nil
	}
	mapped := sqlcgen.ReviewStatus(*status)
	return &mapped
}

func mapReview(row sqlcgen.ListReviewsByOrderRow) domainreview.Review {
	return domainreview.Review{
		ID:        row.ID,
		ProductID: row.ProductID,
		Rating:    row.Rating,
		Comment:   row.Comment,
		Status:    domainreview.Status(row.Status),
		CreatedAt: row.CreatedAt,
	}
}

func mapManagerReview(row sqlcgen.ManagerListReviewsRow) port.ManagerReview {
	return port.ManagerReview{
		ID:            row.ID,
		OrderCode:     row.OrderCode,
		ProductID:     row.ProductID,
		ProductName:   row.ProductName,
		Rating:        row.Rating,
		Comment:       row.Comment,
		Status:        domainreview.Status(row.Status),
		CreatedAt:     row.CreatedAt,
		CustomerName:  row.CustomerName,
		ModeratorName: row.ModeratorName,
	}
}

var _ port.ReviewRepository = (*ReviewRepository)(nil)
