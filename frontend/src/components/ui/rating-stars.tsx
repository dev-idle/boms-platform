import { StarIcon } from "@/components/icons/storefront-icons";
import { MAX_RATING } from "@/lib/schemas/review";
import { cn } from "@/lib/utils";

type RatingStarsProps = {
  /** Whole stars, or an average, which fills the nearest whole number of stars. */
  rating: number;
  className?: string;
};

const STARS = Array.from({ length: MAX_RATING }, (_, index) => index + 1);

/** A rating as stars; the label reads it as a number, so the stars are never the only channel. */
export function RatingStars({ rating, className }: RatingStarsProps) {
  const filled = Math.round(rating);
  return (
    <span aria-label={`${rating} out of ${MAX_RATING} stars`} className={cn("rating-stars", className)} role="img">
      {STARS.map((star) => (
        <span className="rating-stars__star" data-filled={star <= filled ? "true" : undefined} key={star}>
          <StarIcon />
        </span>
      ))}
    </span>
  );
}
