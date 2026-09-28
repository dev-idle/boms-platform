"use client";

import { Input } from "@/components/ui/input";
import { PICKUP_COPY } from "@/constants/pickup";

const HINT_ID = "pickup-at-hint";

type PickupSlotPickerProps = {
  value: string;
  onChange: (value: string) => void;
  /** Earliest and latest pickup the rules allow, as `datetime-local` bounds. */
  min: string;
  max: string;
  /** Empty while there are no rules to describe. */
  hint: string;
  /** The message that says what is wrong with the value, if anything. */
  errorId: string;
  invalid: boolean;
  disabled?: boolean;
};

export function PickupSlotPicker({
  value,
  onChange,
  min,
  max,
  hint,
  errorId,
  invalid,
  disabled = false,
}: PickupSlotPickerProps) {
  return (
    <div className="storefront-pickup-picker">
      <label className="storefront-pickup-picker__label" htmlFor="pickup-at">
        {PICKUP_COPY.label}
      </label>
      {hint ? (
        <p className="storefront-pickup-picker__hint text-caption" id={HINT_ID}>
          {hint}
        </p>
      ) : null}
      <Input
        aria-describedby={hint ? `${HINT_ID} ${errorId}` : errorId}
        aria-invalid={invalid || undefined}
        disabled={disabled}
        id="pickup-at"
        max={max || undefined}
        min={min || undefined}
        required
        type="datetime-local"
        value={value}
        variant="inline"
        onChange={(event) => onChange(event.target.value)}
      />
    </div>
  );
}
