package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainreview "github.com/boms/backend/internal/domain/review"
)

// ReviewRepository keeps customers' reviews of the products they picked up.
type ReviewRepository interface {
	// Create returns apperrors.ErrConflict when the product on the order is
	// already reviewed.
	Create(ctx context.Context, params CreateReviewParams) (*domainreview.Review, error)
	// ListByOrder returns the customer's reviews of the products on their order.
	ListByOrder(ctx context.Context, userID, orderID uuid.UUID) ([]domainreview.Review, error)
	// ListPublished pages a product's published reviews, latest first: the
	// latest, or those written before the review beforeID.
	ListPublished(ctx context.Context, productID uuid.UUID, beforeID *uuid.UUID, limit int32) ([]PublishedReview, error)
	// PublishedRating adds up a product's published reviews.
	PublishedRating(ctx context.Context, productID uuid.UUID) (RatingTotal, error)
	ManagerList(ctx context.Context, params ManagerListReviewsParams) ([]ManagerReview, error)
	ManagerListCount(ctx context.Context, status *domainreview.Status) (int64, error)
	// GetForUpdate holds the review until the transaction ends.
	GetForUpdate(ctx context.Context, id uuid.UUID) (LockedReview, error)
	Moderate(ctx context.Context, id, moderatorID uuid.UUID, status domainreview.Status) (*ManagerReview, error)
	// RatingCounts adds up every review not hidden, and counts those pending
	// and each rating.
	RatingCounts(ctx context.Context) (RatingCounts, error)
	// RatingsByProduct adds up each reviewed product's reviews not hidden, the
	// lowest average first. Both leave out the products the bakery deleted.
	RatingsByProduct(ctx context.Context, limit int32) ([]ProductRating, error)
	// ListForExport returns every review the customer wrote, oldest first.
	ListForExport(ctx context.Context, userID uuid.UUID) ([]ReviewEntry, error)
	// EraseForCustomer erases what the customer wrote; their reviews no longer count.
	EraseForCustomer(ctx context.Context, userID uuid.UUID) error
}

// CreateReviewParams is a customer's checked review of a product on their order.
type CreateReviewParams struct {
	OrderID   uuid.UUID
	ProductID uuid.UUID
	UserID    uuid.UUID
	Rating    int16
	Comment   *string
}

// PublishedReview is a review as the storefront shows it, without its author.
type PublishedReview struct {
	ID        uuid.UUID
	Rating    int16
	Comment   *string
	CreatedAt time.Time
}

// LockedReview is where a review stands and who wrote it.
type LockedReview struct {
	Status   domainreview.Status
	AuthorID uuid.UUID
}

// RatingTotal is how many reviews there are and their ratings added up.
type RatingTotal struct {
	Count int64
	Total int64
}

// RatingCounts adds up reviews, counts those pending and each rating:
// Stars[0] holds the one-star reviews.
type RatingCounts struct {
	Count   int64
	Pending int64
	Total   int64
	Stars   [domainreview.MaxRating]int64
}

// ProductRating adds up one product's reviews.
type ProductRating struct {
	ProductID   uuid.UUID
	ProductName string
	Count       int64
	Total       int64
}

// ManagerListReviewsParams pages the reviews to moderate. Status nil lists
// every review.
type ManagerListReviewsParams struct {
	Status *domainreview.Status
	Limit  int32
	Offset int32
}

// ManagerReview is a review as a manager moderates it: with its order,
// product, the name its author gave and who moderated it last.
type ManagerReview struct {
	ID            uuid.UUID
	OrderCode     string
	ProductID     uuid.UUID
	ProductName   string
	Rating        int16
	Comment       *string
	Status        domainreview.Status
	CreatedAt     time.Time
	CustomerName  *string
	ModeratorName *string
}

// ReviewEntry is what a personal data export holds about one review.
type ReviewEntry struct {
	OrderID     uuid.UUID
	OrderCode   string
	ProductID   uuid.UUID
	ProductName string
	Rating      int16
	Comment     *string
	Status      domainreview.Status
	CreatedAt   time.Time
}
