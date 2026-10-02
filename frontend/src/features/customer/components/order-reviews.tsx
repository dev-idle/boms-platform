"use client";

import { Button } from "@/components/ui/button";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { RatingStars } from "@/components/ui/rating-stars";
import { reviewStatusToPillVariant, StatusPill } from "@/components/ui/status-pill";
import type { ReviewStatus } from "@/lib/schemas/review";

import { useOrderReviews } from "../hooks";
import type { ReviewableProduct } from "../lib/reviewable-products";
import { OrderReviewForm } from "./order-review-form";

type OrderReviewsProps = {
  orderId: string;
  products: ReviewableProduct[];
};

const STATUS_LABEL: Record<ReviewStatus, string> = {
  pending: "Awaiting approval",
  published: "Published",
  hidden: "Not published",
};

/** The products the customer picked up on an order, each reviewed once. */
export function OrderReviews({ orderId, products }: OrderReviewsProps) {
  const reviewsQuery = useOrderReviews(orderId);
  const reviews = new Map(reviewsQuery.data?.map((review) => [review.product_id, review]));

  return (
    <section aria-labelledby="order-reviews" className="storefront-panel storefront-order-reviews">
      <div className="storefront-order-reviews__header">
        <h2 className="text-section-heading" id="order-reviews">
          Reviews
        </h2>
        <p className="text-caption">
          How was it? A manager reads every review before it appears on the product&apos;s page, without your name.
        </p>
      </div>
      {reviewsQuery.isPending ? (
        <InlineLoadingState />
      ) : reviewsQuery.isError ? (
        <div className="storefront-empty-state">
          <p className="storefront-empty-state__message text-error">We could not load your reviews.</p>
          <Button type="button" variant="outline" onClick={() => void reviewsQuery.refetch()}>
            Try again
          </Button>
        </div>
      ) : (
        <ul className="review-list">
          {products.map((product) => {
            const review = reviews.get(product.id);
            return (
              <li className="review-list__item" key={product.id}>
                <p className="text-section-heading">{product.name}</p>
                {review ? (
                  <>
                    <div className="review-list__meta">
                      <RatingStars rating={review.rating} />
                      <StatusPill label={STATUS_LABEL[review.status]} variant={reviewStatusToPillVariant(review.status)} />
                    </div>
                    {review.comment ? <p className="review-comment">{review.comment}</p> : null}
                  </>
                ) : (
                  <OrderReviewForm orderId={orderId} productId={product.id} productName={product.name} />
                )}
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
