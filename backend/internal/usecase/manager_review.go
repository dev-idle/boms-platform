package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	domainreview "github.com/boms/backend/internal/domain/review"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
)

// reviewSummaryProductLimit is how many products the feedback summary lists.
const reviewSummaryProductLimit int32 = 100

// ManagerReviewUsecase is a manager reading what customers think and deciding
// which reviews the storefront shows.
type ManagerReviewUsecase struct {
	reviews port.ReviewRepository
	tx      port.TxManager
	events  port.EventOutbox
	audit   *auditlogger.Service
	log     *zap.Logger
}

func NewManagerReviewUsecase(
	reviews port.ReviewRepository,
	tx port.TxManager,
	events port.EventOutbox,
	audit *auditlogger.Service,
	log *zap.Logger,
) *ManagerReviewUsecase {
	return &ManagerReviewUsecase{reviews: reviews, tx: tx, events: events, audit: audit, log: log}
}

// List returns a page of reviews, latest first. statusFilter narrows it to
// pending, published or hidden reviews.
func (u *ManagerReviewUsecase) List(
	ctx context.Context,
	page, pageSize int32,
	statusFilter string,
) ([]dto.ManagerReviewResponse, int64, int32, int32, error) {
	page, pageSize = normalizeCatalogListPage(page, pageSize)
	var status *domainreview.Status
	if trimmed := strings.TrimSpace(statusFilter); trimmed != "" {
		parsed := domainreview.Status(trimmed)
		if !parsed.Valid() {
			return nil, 0, page, pageSize, apperrors.ErrValidation.WithDetail("status", "must be pending, published or hidden")
		}
		status = &parsed
	}
	rows, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]port.ManagerReview, error) {
			return u.reviews.ManagerList(ctx, port.ManagerListReviewsParams{
				Status: status,
				Limit:  pageSize,
				Offset: utils.PageOffset(page, pageSize),
			})
		},
		func(ctx context.Context) (int64, error) {
			return u.reviews.ManagerListCount(ctx, status)
		},
	)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]dto.ManagerReviewResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toManagerReviewResponse(row))
	}
	return out, total, page, pageSize, nil
}

// Summary adds up every review not hidden of a product the bakery keeps:
// overall, by rating and by product. A pending review counts — it is a
// customer's word on a pickup — but one a manager hid does not.
func (u *ManagerReviewUsecase) Summary(ctx context.Context) (*dto.ReviewSummaryResponse, error) {
	var counts port.RatingCounts
	var products []port.ProductRating
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() (err error) {
		counts, err = u.reviews.RatingCounts(groupCtx)
		return err
	})
	group.Go(func() (err error) {
		products, err = u.reviews.RatingsByProduct(groupCtx, reviewSummaryProductLimit)
		return err
	})
	if err := group.Wait(); err != nil {
		return nil, err
	}
	out := make([]dto.ProductRatingResponse, 0, len(products))
	for _, product := range products {
		out = append(out, dto.ProductRatingResponse{
			ProductID:     product.ProductID.String(),
			ProductName:   product.ProductName,
			ReviewCount:   product.Count,
			AverageRating: *averageRating(product.Count, product.Total),
		})
	}
	return &dto.ReviewSummaryResponse{
		ReviewCount:   counts.Count,
		PendingCount:  counts.Pending,
		AverageRating: averageRating(counts.Count, counts.Total),
		Stars:         counts.Stars[:],
		Products:      out,
	}, nil
}

// Moderate publishes a review on the storefront or hides it; managers and the
// author hear of it.
func (u *ManagerReviewUsecase) Moderate(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	reviewID uuid.UUID,
	req dto.ModerateReviewRequest,
) (*dto.ManagerReviewResponse, error) {
	status, err := domainreview.NewModeration(req.Status)
	if err != nil {
		return nil, err
	}
	var before port.LockedReview
	var moderated *port.ManagerReview
	err = u.tx.WithTx(ctx, func(txCtx context.Context) (err error) {
		if before, err = u.reviews.GetForUpdate(txCtx, reviewID); err != nil {
			return err
		}
		if moderated, err = u.reviews.Moderate(txCtx, reviewID, actorID, status); err != nil {
			return err
		}
		return u.events.Add(txCtx, domainreview.ChangedEvent(moderated.ProductID, before.AuthorID))
	})
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, domainreview.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	recordAudit(u.log, u.audit, ctx, domainreview.AuditActionManagerModerated, actorID, actorRole, &reviewID, "review",
		map[string]string{"status": string(before.Status)}, map[string]string{"status": string(status)})
	resp := toManagerReviewResponse(*moderated)
	return &resp, nil
}

func toManagerReviewResponse(review port.ManagerReview) dto.ManagerReviewResponse {
	return dto.ManagerReviewResponse{
		ID:            review.ID.String(),
		OrderCode:     review.OrderCode,
		ProductName:   review.ProductName,
		Rating:        review.Rating,
		Comment:       review.Comment,
		Status:        string(review.Status),
		CreatedAt:     review.CreatedAt,
		CustomerName:  review.CustomerName,
		ModeratorName: review.ModeratorName,
	}
}
