"use client";

import { useMemo } from "react";

import { Input } from "@/components/ui/input";
import { PICKUP_COPY } from "@/constants/pickup";
import {
  maxPickupLocalInputValue,
  minPickupLocalInputValue,
} from "@/lib/validation/pickup";

type PickupSlotPickerProps = {
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
};

export function PickupSlotPicker({
  value,
  onChange,
  disabled = false,
}: PickupSlotPickerProps) {
  // Bounds are anchored to mount time; server-side validation remains authoritative.
  const bounds = useMemo(
    () => ({ min: minPickupLocalInputValue(), max: maxPickupLocalInputValue() }),
    [],
  );

  return (
    <div className="storefront-pickup-picker">
      <label className="storefront-pickup-picker__label" htmlFor="pickup-at">
        {PICKUP_COPY.label}
      </label>
      <p className="storefront-pickup-picker__hint text-caption">{PICKUP_COPY.hint}</p>
      <Input
        className="storefront-pickup-picker__input"
        disabled={disabled}
        id="pickup-at"
        max={bounds.max}
        min={bounds.min}
        required
        type="datetime-local"
        value={value}
        variant="inline"
        onChange={(event) => onChange(event.target.value)}
      />
    </div>
  );
}
