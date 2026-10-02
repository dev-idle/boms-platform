package dto

import "time"

// CreateReviewRequest is a customer's review of a product on their order as
// written; the usecase checks the rating and the comment.
type CreateReviewRequest struct {
	ProductID string  `json:"product_id" validate:"required,uuid"`
	Rating    int     `json:"rating"`
	Comment   *string `json:"comment"`
}

// ReviewResponse is the customer's review of a product on their order, with
// where it stands with the manager.
type ReviewResponse struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	Rating    int16     `json:"rating"`
	Comment   *string   `json:"comment"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ProductReviewsResponse is a page of a product's published reviews, latest
// first, with how many it has and their average rating, null before the
// first. HasMore says older ones exist; the next page starts before the last.
type ProductReviewsResponse struct {
	ReviewCount   int64                     `json:"review_count"`
	AverageRating *float64                  `json:"average_rating"`
	Reviews       []PublishedReviewResponse `json:"reviews"`
	HasMore       bool                      `json:"has_more"`
}

// PublishedReviewResponse is a review as the storefront shows it: from a
// customer who picked the product up, without their name.
type PublishedReviewResponse struct {
	ID        string    `json:"id"`
	Rating    int16     `json:"rating"`
	Comment   *string   `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

// ManagerReviewResponse is a review as a manager moderates it. ModeratorName
// is null while it is pending.
type ManagerReviewResponse struct {
	ID            string    `json:"id"`
	OrderCode     string    `json:"order_code"`
	ProductName   string    `json:"product_name"`
	Rating        int16     `json:"rating"`
	Comment       *string   `json:"comment"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	CustomerName  *string   `json:"customer_name"`
	ModeratorName *string   `json:"moderator_name"`
}

// ModerateReviewRequest publishes a review or hides it.
type ModerateReviewRequest struct {
	Status string `json:"status" validate:"required,oneof=published hidden"`
}

// ReviewSummaryResponse is the customers' feedback for a manager: every
// review not hidden, how many wait for a manager, how many gave each rating
// (Stars[0] one star), and each product's, the lowest average first.
type ReviewSummaryResponse struct {
	ReviewCount   int64                   `json:"review_count"`
	PendingCount  int64                   `json:"pending_count"`
	AverageRating *float64                `json:"average_rating"`
	Stars         []int64                 `json:"stars"`
	Products      []ProductRatingResponse `json:"products"`
}

// ProductRatingResponse is one product's reviews added up.
type ProductRatingResponse struct {
	ProductID     string  `json:"product_id"`
	ProductName   string  `json:"product_name"`
	ReviewCount   int64   `json:"review_count"`
	AverageRating float64 `json:"average_rating"`
}
