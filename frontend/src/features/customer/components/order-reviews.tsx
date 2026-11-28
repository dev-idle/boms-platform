"use client";

import { StorefrontErrorState } from "@/components/layouts/storefront-state";
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
  published: "Published",
  hidden: "Hidden by the bakery",
};

/** The products the customer picked up on an order, each reviewed once. */
export function OrderReviews({ orderId, products }: OrderReviewsProps) {
  const reviewsQuery = useOrderReviews(orderId);
  const reviews = new Map(reviewsQuery.data?.map((review) => [review.product_id, review]));

  return (
    <section aria-labelledby="order-reviews" className="storefront-order-section">
      <div className="storefront-order-section__header">
        <h2 className="storefront-section-heading" id="order-reviews">
          Reviews
        </h2>
        <p className="text-caption">
          A review appears on the product&apos;s page right away, signed with your first name and last initial if your
          account has a name.
        </p>
      </div>
      {reviewsQuery.isPending ? (
        <div aria-busy="true" className="grid gap-3">
          <span className="skeleton block h-5 w-48" />
          <span className="skeleton block h-4 w-full max-w-md" />
        </div>
      ) : reviewsQuery.isError ? (
        <StorefrontErrorState onRetry={() => void reviewsQuery.refetch()} title="We could not load your reviews" />
      ) : (
        <ul className="review-list">
          {products.map((product) => {
            const review = reviews.get(product.id);
            return (
              <li className="review-list__item" key={product.id}>
                <p className="storefront-line-name">{product.name}</p>
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
