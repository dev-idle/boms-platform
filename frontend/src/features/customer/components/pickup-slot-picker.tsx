"use client";

import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { PICKUP_COPY } from "@/constants/pickup";

const HINT_ID = "pickup-hint";
const TYPE_ID = "pickup-type";

export type PickupSlotOption = {
  /** The slot start as a bakery-local `datetime-local` value. */
  value: string;
  /** The time, with the reason when checkout would refuse it. */
  label: string;
  /** The slot holds as many orders as it takes. */
  full: boolean;
  /** Checkout would refuse it: full, or too soon for these items. */
  disabled: boolean;
};

type PickupSlotPickerProps = {
  day: string;
  onDayChange: (day: string) => void;
  /** First and last bakery day these items can be collected on (YYYY-MM-DD). */
  minDay: string;
  maxDay: string;
  slot: string;
  onSlotChange: (slot: string) => void;
  slots: PickupSlotOption[];
  /** The day's slots are on their way. */
  loadingSlots: boolean;
  /** Empty while there are no rules to describe. */
  hint: string;
  /** How an order at the chosen slot is prepared; empty when none is chosen. */
  typeNote: string;
  /** The message that says what is wrong with the choice, if anything. */
  errorId: string;
  dayInvalid: boolean;
  slotInvalid: boolean;
  disabled?: boolean;
};

/** A pickup day, then one of that day's slots; full and too-soon slots are greyed with the reason. */
export function PickupSlotPicker({
  day,
  onDayChange,
  minDay,
  maxDay,
  slot,
  onSlotChange,
  slots,
  loadingSlots,
  hint,
  typeNote,
  errorId,
  dayInvalid,
  slotInvalid,
  disabled = false,
}: PickupSlotPickerProps) {
  const describedBy = [hint ? HINT_ID : null, typeNote ? TYPE_ID : null, errorId].filter(Boolean).join(" ");
  return (
    <div className="storefront-pickup-picker">
      {hint ? (
        <p className="storefront-pickup-picker__hint text-caption" id={HINT_ID}>
          {hint}
        </p>
      ) : null}
      <div className="storefront-pickup-picker__fields">
        <div className="storefront-pickup-picker__field">
          <label className="storefront-pickup-picker__label" htmlFor="pickup-day">
            {PICKUP_COPY.dayLabel}
          </label>
          <Input
            aria-describedby={describedBy}
            aria-invalid={dayInvalid || undefined}
            disabled={disabled}
            id="pickup-day"
            max={maxDay || undefined}
            min={minDay || undefined}
            required
            type="date"
            value={day}
            variant="inline"
            onChange={(event) => onDayChange(event.target.value)}
          />
        </div>
        <div className="storefront-pickup-picker__field">
          <label className="storefront-pickup-picker__label" htmlFor="pickup-slot">
            {PICKUP_COPY.label}
          </label>
          <Select
            aria-busy={loadingSlots || undefined}
            aria-describedby={describedBy}
            aria-invalid={slotInvalid || undefined}
            className="field-chrome--inline"
            contentClassName="storefront-pickup-picker__slots"
            disabled={disabled || slots.length === 0}
            id="pickup-slot"
            required
            value={slot}
            onChange={(event) => onSlotChange(event.target.value)}
          >
            <option value="">{loadingSlots ? PICKUP_COPY.loading : PICKUP_COPY.slotPlaceholder}</option>
            {slots.map((option) => (
              <option key={option.value} disabled={option.disabled} value={option.value}>
                {option.label}
              </option>
            ))}
          </Select>
        </div>
      </div>
      {typeNote ? (
        <p className="storefront-pickup-picker__type text-caption" id={TYPE_ID}>
          {typeNote}
        </p>
      ) : null}
    </div>
  );
}
