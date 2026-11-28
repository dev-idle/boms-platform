"use client";

import { StorefrontErrorState } from "@/components/layouts/storefront-state";
import { Button } from "@/components/ui/button";
import { RatingStars } from "@/components/ui/rating-stars";
import { formatAverageRating } from "@/lib/schemas/review";
import { splitDateTime } from "@/lib/validation/datetime";

import { useProductReviews } from "../hooks";

type ProductReviewsProps = {
  productId: string;
};

/** What customers who picked the product up said of it, each under the short form of their name. */
export function ProductReviews({ productId }: ProductReviewsProps) {
  const reviewsQuery = useProductReviews(productId);
  const pages = reviewsQuery.data?.pages ?? [];
  const summary = pages[0];
  const reviews = pages.flatMap((page) => page.reviews);

  return (
    <section aria-labelledby="product-reviews" className="catalog-reviews">
      <div className="catalog-reviews__header">
        <h2 className="storefront-section-heading" id="product-reviews">
          Reviews
        </h2>
        {summary?.average_rating ? (
          <p className="catalog-reviews__summary">
            <span className="catalog-reviews__average">{formatAverageRating(summary.average_rating)}</span>
            <RatingStars rating={summary.average_rating} />
            <span className="text-caption">
              {summary.review_count} {summary.review_count === 1 ? "review" : "reviews"}
            </span>
          </p>
        ) : null}
      </div>
      {reviewsQuery.isPending ? (
        <div aria-busy="true" className="grid gap-3">
          <span className="skeleton block h-4 w-40" />
          <span className="skeleton block h-4 w-full max-w-md" />
        </div>
      ) : reviewsQuery.isError ? (
        <StorefrontErrorState onRetry={() => void reviewsQuery.refetch()} title="We could not load the reviews" />
      ) : reviews.length === 0 ? (
        <p className="text-caption">No reviews yet. Customers review what they picked up.</p>
      ) : (
        <>
          <ul className="review-list">
            {reviews.map((review) => (
              <li className="review-list__item" key={review.id}>
                <div className="review-list__meta">
                  <RatingStars rating={review.rating} />
                  <span className="text-caption">
                    {review.author_name ? `${review.author_name} · ` : null}Verified pickup ·{" "}
                    {splitDateTime(review.created_at).date}
                  </span>
                </div>
                {review.comment ? <p className="review-comment">{review.comment}</p> : null}
              </li>
            ))}
          </ul>
          {reviewsQuery.hasNextPage ? (
            <Button
              aria-busy={reviewsQuery.isFetchingNextPage || undefined}
              disabled={reviewsQuery.isFetchingNextPage}
              type="button"
              variant="outline"
              onClick={() => void reviewsQuery.fetchNextPage()}
            >
              {reviewsQuery.isFetchingNextPage ? "Loading…" : "Show more reviews"}
            </Button>
          ) : null}
        </>
      )}
    </section>
  );
}
