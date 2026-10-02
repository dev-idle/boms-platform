import { z } from "zod";

import { apiDateTimeSchema } from "@/lib/validation/datetime";

/** The most stars a rating gives — mirrors backend `review.MaxRating`. */
export const MAX_RATING = 5;

/**
 * The most characters a review's comment holds — mirrors backend
 * `review.MaxCommentLength`, which counts characters (code points).
 */
export const REVIEW_COMMENT_MAX_LENGTH = 1000;

/** Where a review stands with the manager — mirrors backend `review.Status`. */
export const reviewStatusSchema = z.enum(["pending", "published", "hidden"]);

export type ReviewStatus = z.infer<typeof reviewStatusSchema>;

export const ratingSchema = z.number().int().min(1).max(MAX_RATING);

/** An average rating to one decimal, as the API sends it. */
export const averageRatingSchema = z.number().min(1).max(MAX_RATING);

/**
 * GET /catalog/products/:id/reviews — a page of a product's published reviews,
 * latest first, with how many it has and their average, null before the first.
 * A review shows no author: everyone who wrote one picked the product up.
 */
export const productReviewsSchema = z.object({
  review_count: z.number().int().min(0),
  average_rating: averageRatingSchema.nullable(),
  reviews: z.array(
    z.object({
      id: z.uuid(),
      rating: ratingSchema,
      comment: z.string().min(1).nullable(),
      created_at: apiDateTimeSchema,
    }),
  ),
  has_more: z.boolean(),
});

export type ProductReviews = z.infer<typeof productReviewsSchema>;

/** An average rating as it is read: "4.5". */
export function formatAverageRating(average: number): string {
  return average.toFixed(1);
}
