"use client";

import { useState } from "react";

import { PICKUP_COPY } from "@/constants/pickup";
import { cn } from "@/lib/utils";
import {
  bakeryPickupISOFromLocalInput,
  formatPickupDayChoice,
  formatPickupWallTime,
  pickupDaysBetween,
} from "@/lib/validation/pickup";

const HINT_ID = "pickup-hint";
const TYPE_ID = "pickup-type";
const DAY_LABEL_ID = "pickup-day-label";
const SLOT_LABEL_ID = "pickup-slot-label";

/** Parts of the bakery day, each up to the hour it ends. */
const PERIODS = [
  { key: "morning", label: "Morning", endsAtHour: 12 },
  { key: "afternoon", label: "Afternoon", endsAtHour: 17 },
  { key: "evening", label: "Evening", endsAtHour: 24 },
] as const;

type Period = (typeof PERIODS)[number]["key"];

/** The part of the day a bakery-local slot value starts in. */
function periodOf(value: string): Period {
  const hour = Number(value.slice(11, 13));
  return (PERIODS.find((period) => hour < period.endsAtHour) ?? PERIODS[2]).key;
}

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

/**
 * A row of the days these items can be collected on, then the chosen day's
 * slots as buttons, one part of the day at a time so a long day stays short;
 * a full or too-soon slot is greyed and struck through, its reason in its name.
 */
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
  const days = pickupDaysBetween(minDay, maxDay);
  // The part of the day on show: the one asked for, else the chosen slot's,
  // else the first with a time left, else the first the day has.
  const [askedPeriod, setAskedPeriod] = useState<Period | null>(null);
  const periods = PERIODS.filter((period) => slots.some((option) => periodOf(option.value) === period.key));
  const has = (key: Period | null): key is Period => periods.some((period) => period.key === key);
  const chosenPeriod = slot ? periodOf(slot) : null;
  const openPeriod = periods.find((period) =>
    slots.some((option) => !option.disabled && periodOf(option.value) === period.key),
  )?.key;
  const shownPeriod = has(askedPeriod) ? askedPeriod : has(chosenPeriod) ? chosenPeriod : (openPeriod ?? periods[0]?.key);
  const shownSlots = slots.filter((option) => periodOf(option.value) === shownPeriod);

  return (
    <div className="storefront-pickup-picker">
      {hint ? (
        <p className="storefront-pickup-picker__hint text-caption" id={HINT_ID}>
          {hint}
        </p>
      ) : null}

      <div className="storefront-pickup-picker__field">
        <p className="storefront-pickup-picker__label text-form-label" id={DAY_LABEL_ID}>
          {PICKUP_COPY.dayLabel}
        </p>
        <div
          aria-describedby={describedBy}
          aria-labelledby={DAY_LABEL_ID}
          className={cn("pickup-days", dayInvalid && "pickup-days--invalid")}
          role="group"
        >
          {days.map((choice, index) => {
            const { weekday, date, month, monthLong } = formatPickupDayChoice(choice);
            const selected = choice === day;
            // The month shows on the first day and on each 1st; elsewhere only
            // assistive technology reads it.
            const showMonth = index === 0 || date === "1";
            return (
              <button
                aria-pressed={selected}
                className={cn("pickup-day", selected && "pickup-day--selected")}
                disabled={disabled}
                key={choice}
                onClick={() => onDayChange(choice)}
                type="button"
              >
                <span className="pickup-day__weekday">{weekday}</span>
                <span className="pickup-day__date">{showMonth ? `${month} ${date}` : date}</span>
                {showMonth ? null : <span className="sr-only"> {monthLong}</span>}
              </button>
            );
          })}
        </div>
      </div>

      <div className="storefront-pickup-picker__field">
        <p className="storefront-pickup-picker__label text-form-label" id={SLOT_LABEL_ID}>
          {PICKUP_COPY.label}
        </p>
        {slots.length === 0 ? (
          <p aria-busy={loadingSlots || undefined} className="text-caption">
            {loadingSlots ? PICKUP_COPY.loading : PICKUP_COPY.slotPlaceholder}
          </p>
        ) : (
          <>
            {periods.length > 1 ? (
              <div aria-label="Part of the day" className="pickup-periods" role="group">
                {periods.map((period) => {
                  const open = slots.filter((option) => !option.disabled && periodOf(option.value) === period.key).length;
                  return (
                    <button
                      aria-pressed={period.key === shownPeriod}
                      className="pickup-period"
                      disabled={disabled}
                      key={period.key}
                      onClick={() => setAskedPeriod(period.key)}
                      type="button"
                    >
                      {period.label}
                      <span className="pickup-period__count">
                        {open}
                        <span className="sr-only"> {open === 1 ? "time" : "times"} left</span>
                      </span>
                    </button>
                  );
                })}
              </div>
            ) : null}
            <div
              aria-describedby={describedBy}
              aria-labelledby={SLOT_LABEL_ID}
              className={cn("pickup-slots", slotInvalid && "pickup-slots--invalid")}
              role="group"
            >
              {shownSlots.map((option) => {
                const selected = option.value === slot;
                return (
                  <button
                    aria-label={option.label}
                    aria-pressed={selected}
                    className={cn("pickup-slot", selected && "pickup-slot--selected")}
                    disabled={disabled || option.disabled}
                    key={option.value}
                    onClick={() => onSlotChange(option.value)}
                    title={option.disabled ? option.label : undefined}
                    type="button"
                  >
                    {formatPickupWallTime(bakeryPickupISOFromLocalInput(option.value))}
                  </button>
                );
              })}
            </div>
          </>
        )}
      </div>

      {typeNote ? (
        <p className="storefront-pickup-picker__type text-caption" id={TYPE_ID}>
          {typeNote}
        </p>
      ) : null}
    </div>
  );
}
