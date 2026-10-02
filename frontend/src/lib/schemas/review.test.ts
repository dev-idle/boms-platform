import { describe, expect, it } from "vitest";

import { formatAverageRating, productReviewsSchema } from "./review";

describe("productReviewsSchema", () => {
  const page = {
    review_count: 2,
    average_rating: 4.5,
    reviews: [
      { id: "0b6c8f5e-3c1d-4a8e-9f0a-2d7e6c5b4a39", rating: 5, comment: "Flaky\nand buttery", created_at: "2026-10-02T09:00:00+07:00" },
      { id: "1c7d9a6f-4d2e-4b9f-8a1b-3e8f7d6c5b4a", rating: 4, comment: null, created_at: "2026-10-01T09:00:00+07:00" },
    ],
    has_more: false,
  };

  it("reads a page of published reviews with their average", () => {
    expect(productReviewsSchema.parse(page).reviews).toHaveLength(2);
  });

  it("reads a product nobody has reviewed yet", () => {
    expect(productReviewsSchema.safeParse({ review_count: 0, average_rating: null, reviews: [], has_more: false }).success).toBe(
      true,
    );
  });

  it("refuses a rating outside one to five stars", () => {
    const bad = { ...page, reviews: [{ ...page.reviews[0], rating: 6 }] };
    expect(productReviewsSchema.safeParse(bad).success).toBe(false);
  });
});

describe("formatAverageRating", () => {
  it("reads an average to one decimal", () => {
    expect(formatAverageRating(4)).toBe("4.0");
    expect(formatAverageRating(4.5)).toBe("4.5");
  });
});
