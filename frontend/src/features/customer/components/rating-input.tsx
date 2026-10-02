"use client";

import { StarIcon } from "@/components/icons/storefront-icons";
import { MAX_RATING } from "@/lib/schemas/review";

type RatingInputProps = {
  /** Groups the stars into one choice. */
  name: string;
  value: number;
  disabled: boolean;
  onChange: (rating: number) => void;
};

const STARS = Array.from({ length: MAX_RATING }, (_, index) => index + 1);

/** Stars to choose a rating from: a radio each, filled up to the one chosen. */
export function RatingInput({ name, value, disabled, onChange }: RatingInputProps) {
  return (
    <div className="rating-input">
      {STARS.map((star) => (
        <label className="rating-input__star" data-filled={star <= value ? "true" : undefined} key={star}>
          <input
            checked={value === star}
            className="sr-only"
            disabled={disabled}
            name={name}
            type="radio"
            value={star}
            onChange={() => onChange(star)}
          />
          <StarIcon />
          <span className="sr-only">{star === 1 ? "1 star" : `${star} stars`}</span>
        </label>
      ))}
    </div>
  );
}
