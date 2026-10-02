package usecase

import (
	"context"
	"errors"
	"math"
	"slices"

	"github.com/google/uuid"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainreview "github.com/boms/backend/internal/domain/review"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// productReviewPageSize is how many reviews one page of a product's holds.
const productReviewPageSize int32 = 10

// ReviewUsecase is a customer reviewing the products they picked up, and the
// storefront showing the reviews a manager published.
type ReviewUsecase struct {
	users   port.UserRepository
	orders  port.OrderRepository
	reviews port.ReviewRepository
	tx      port.TxManager
	events  port.EventOutbox
}

func NewReviewUsecase(
	users port.UserRepository,
	orders port.OrderRepository,
	reviews port.ReviewRepository,
	tx port.TxManager,
	events port.EventOutbox,
) *ReviewUsecase {
	return &ReviewUsecase{users: users, orders: orders, reviews: reviews, tx: tx, events: events}
}

// ListByOrder returns the customer's reviews of the products on their order.
func (u *ReviewUsecase) ListByOrder(ctx context.Context, userID, orderID uuid.UUID) ([]dto.ReviewResponse, error) {
	reviews, err := u.reviews.ListByOrder(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ReviewResponse, 0, len(reviews))
	for _, review := range reviews {
		out = append(out, toReviewResponse(review))
	}
	return out, nil
}

// Create records the customer's review of a product they picked up on their
// order: one per product on the order. It waits for a manager, who hears of
// it at once.
func (u *ReviewUsecase) Create(ctx context.Context, userID, orderID uuid.UUID, req dto.CreateReviewRequest) (*dto.ReviewResponse, error) {
	rating, err := domainreview.NewRating(req.Rating)
	if err != nil {
		return nil, err
	}
	comment, err := domainreview.NewComment(req.Comment)
	if err != nil {
		return nil, err
	}
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, apperrors.ErrValidation.WithDetail("product_id", "invalid product id")
	}
	if err := u.checkReviewable(ctx, userID, orderID, productID); err != nil {
		return nil, err
	}
	var review *domainreview.Review
	err = u.tx.WithTx(ctx, func(txCtx context.Context) (err error) {
		if err := holdCustomer(txCtx, u.users, userID); err != nil {
			return err
		}
		review, err = u.reviews.Create(txCtx, port.CreateReviewParams{
			OrderID:   orderID,
			ProductID: productID,
			UserID:    userID,
			Rating:    rating,
			Comment:   comment,
		})
		if err != nil {
			return err
		}
		return u.events.Add(txCtx, domainreview.ChangedEvent(productID, userID))
	})
	if errors.Is(err, apperrors.ErrConflict) {
		return nil, domainreview.ErrExists
	}
	if err != nil {
		return nil, err
	}
	resp := toReviewResponse(*review)
	return &resp, nil
}

// checkReviewable confirms the customer picked productID up on their order.
func (u *ReviewUsecase) checkReviewable(ctx context.Context, userID, orderID, productID uuid.UUID) error {
	order, err := u.orders.GetByIDForUser(ctx, userID, orderID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return domainorder.ErrNotFound
	}
	if err != nil {
		return err
	}
	if order.Status != domainorder.StatusFulfilled {
		return domainreview.ErrNotReviewable
	}
	items, err := u.orders.ListItemsByOrderID(ctx, orderID)
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(items, func(item domainorder.Item) bool {
		return item.ProductID != nil && *item.ProductID == productID
	}) {
		return domainreview.ErrNotReviewable
	}
	return nil
}

// ProductReviews returns a page of a product's published reviews, latest
// first — the latest, or those written before beforeID — with how many it has
// and their average rating.
func (u *ReviewUsecase) ProductReviews(ctx context.Context, productID uuid.UUID, beforeID *uuid.UUID) (*dto.ProductReviewsResponse, error) {
	var rating port.RatingTotal
	reviews, _, err := listWithTotal(ctx,
		func(ctx context.Context) ([]port.PublishedReview, error) {
			return u.reviews.ListPublished(ctx, productID, beforeID, productReviewPageSize+1)
		},
		func(ctx context.Context) (int64, error) {
			var err error
			rating, err = u.reviews.PublishedRating(ctx, productID)
			return rating.Count, err
		},
	)
	if err != nil {
		return nil, err
	}
	hasMore := len(reviews) > int(productReviewPageSize)
	if hasMore {
		reviews = reviews[:productReviewPageSize]
	}
	out := make([]dto.PublishedReviewResponse, 0, len(reviews))
	for _, review := range reviews {
		out = append(out, dto.PublishedReviewResponse{
			ID:        review.ID.String(),
			Rating:    review.Rating,
			Comment:   review.Comment,
			CreatedAt: review.CreatedAt,
		})
	}
	return &dto.ProductReviewsResponse{
		ReviewCount:   rating.Count,
		AverageRating: averageRating(rating.Count, rating.Total),
		Reviews:       out,
		HasMore:       hasMore,
	}, nil
}

func toReviewResponse(review domainreview.Review) dto.ReviewResponse {
	return dto.ReviewResponse{
		ID:        review.ID.String(),
		ProductID: review.ProductID.String(),
		Rating:    review.Rating,
		Comment:   review.Comment,
		Status:    string(review.Status),
		CreatedAt: review.CreatedAt,
	}
}

// averageRating is total over count to one decimal, as the stars are shown;
// nil before the first review.
func averageRating(count, total int64) *float64 {
	if count == 0 {
		return nil
	}
	average := math.Round(float64(total)*10/float64(count)) / 10
	return &average
}
