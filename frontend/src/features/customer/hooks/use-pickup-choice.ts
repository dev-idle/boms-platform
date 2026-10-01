"use client";

import { useState } from "react";

import { PICKUP_COPY } from "@/constants/pickup";
import { useNow } from "@/lib/hooks/use-now";
import type { Fulfillment } from "@/lib/schemas/order";
import { clockToMinutes, formatClockMinutes } from "@/lib/validation/clock";
import {
  bakeryDayOf,
  bakeryPickupISOFromLocalInput,
  earliestPickupLocalValue,
  formatPickupLocalInputValue,
  lastPickupDay,
  pickupOrderType,
  pickupProblem,
  type PickupFulfillment,
  type PickupWindow,
} from "@/lib/validation/pickup";

import type { PickupSlotOption } from "../components/pickup-slot-picker";
import {
  orderTypeNote,
  pickupHint,
  pickupProblemMessage,
  toPickupFulfillment,
  toPickupWindow,
} from "../lib/pickup-window";
import type { PickupSlots } from "../schemas";
import { usePickupRules, usePickupSlots } from "./index";

/** How often the suggested pickup and its checks move with the clock. */
const CLOCK_TICK_MS = 30_000;

/**
 * The day's slots still ahead, as options: a full slot, or one too soon for
 * these items, stays in the list, greyed with the reason.
 */
function slotOptions(
  slots: PickupSlots["slots"],
  pickupWindow: PickupWindow,
  now: Date,
  items: PickupFulfillment,
): PickupSlotOption[] {
  return slots
    .filter(({ starts_at }) => new Date(starts_at).getTime() > now.getTime())
    .map(({ starts_at, full }) => {
      const value = formatPickupLocalInputValue(new Date(starts_at));
      const time = formatClockMinutes(clockToMinutes(value.slice(11)));
      const tooSoon = pickupProblem(value, pickupWindow, now, items) !== null;
      const reason = full ? PICKUP_COPY.full : tooSoon ? PICKUP_COPY.tooSoon : null;
      return { value, label: reason ? `${time} · ${reason}` : time, full, disabled: full || tooSoon };
    });
}

/**
 * A pickup day and slot for items asking `fulfillment` of the bakery, checked
 * against the rules read from the API (refreshed live when an admin edits
 * them) and the day's slots as they fill and free up. Until the customer
 * chooses, the earliest slot that suits the items is offered. `pending` holds
 * the fields while the choice is being sent.
 */
export function usePickupChoice(fulfillment: Fulfillment, pending: boolean) {
  const rulesQuery = usePickupRules();
  const now = useNow(CLOCK_TICK_MS);
  const [chosenDay, setChosenDay] = useState<string | null>(null);
  const [chosenSlot, setChosenSlot] = useState<string | null>(null);

  const items = toPickupFulfillment(fulfillment);
  const pickupWindow = rulesQuery.data ? toPickupWindow(rulesQuery.data) : null;
  const earliest = pickupWindow ? earliestPickupLocalValue(pickupWindow, now, items) : "";
  const minDay = earliest.slice(0, 10) || bakeryDayOf(now);
  const maxDay = pickupWindow ? lastPickupDay(pickupWindow, now) : "";
  const day = chosenDay ?? earliest.slice(0, 10);
  const closedReason = pickupWindow?.closedDates.get(day);

  function dayProblem(): string | null {
    if (!pickupWindow) {
      return rulesQuery.isError ? PICKUP_COPY.unavailable : null;
    }
    if (earliest === "") {
      return PICKUP_COPY.noOpenTime;
    }
    if (day === "") {
      return PICKUP_COPY.required;
    }
    if (day < minDay) {
      return pickupProblemMessage("too_soon", pickupWindow, items, pickupOrderType(earliest, now, items) ?? "pre_order");
    }
    if (day > maxDay) {
      return pickupProblemMessage("too_far", pickupWindow, items, "pre_order");
    }
    if (closedReason !== undefined) {
      return pickupProblemMessage("closed_day", pickupWindow, items, "pre_order", closedReason);
    }
    if (day === items.soldOutOn) {
      return pickupProblemMessage("sold_out", pickupWindow, items, "pre_order");
    }
    return null;
  }

  const dayMessage = dayProblem();
  const slotsQuery = usePickupSlots(pickupWindow && dayMessage === null ? day : "");
  const loadingSlots = dayMessage === null && slotsQuery.isFetching && slotsQuery.data?.date !== day;
  const slots =
    pickupWindow && dayMessage === null && slotsQuery.data?.date === day
      ? slotOptions(slotsQuery.data.slots, pickupWindow, now, items)
      : [];
  const slot = chosenSlot ?? slots.find((option) => !option.disabled)?.value ?? "";

  function slotProblem(): string | null {
    if (!pickupWindow || dayMessage !== null || loadingSlots) {
      return null;
    }
    if (slotsQuery.isError) {
      return PICKUP_COPY.slotsUnavailable;
    }
    if (slot === "") {
      return slotsQuery.isSuccess ? PICKUP_COPY.noSlotLeft : null;
    }
    if (slots.some((option) => option.value === slot && option.full)) {
      return PICKUP_COPY.slotFull;
    }
    const problem = pickupProblem(slot, pickupWindow, now, items);
    if (problem === null) {
      return null;
    }
    if (problem === "missing") {
      return PICKUP_COPY.required;
    }
    return pickupProblemMessage(problem, pickupWindow, items, pickupOrderType(slot, now, items) ?? "pre_order");
  }

  const slotMessage = slotProblem();
  const message = dayMessage ?? slotMessage;
  const type = pickupWindow && slot !== "" && message === null ? pickupOrderType(slot, now, items) : null;

  return {
    /** Everything the picker shows but the id of the message that explains it. */
    picker: {
      day,
      dayInvalid: dayMessage !== null && pickupWindow !== null,
      disabled: pickupWindow === null || pending,
      hint: pickupWindow ? pickupHint(pickupWindow) : rulesQuery.isError ? "" : PICKUP_COPY.loading,
      loadingSlots,
      maxDay,
      minDay,
      slot,
      slotInvalid: slotMessage !== null,
      slots,
      typeNote: pickupWindow && type ? orderTypeNote(type, pickupWindow, items) : "",
      onDayChange: (next: string) => {
        setChosenDay(next);
        setChosenSlot(null);
      },
      onSlotChange: setChosenSlot,
    },
    /** What is wrong with the choice, if anything. */
    message,
    /** A slot is chosen and nothing is wrong with it. */
    ready: pickupWindow !== null && slot !== "" && message === null && !loadingSlots,
    /** Reads the rules or the day's slots again after they failed to load. */
    retry:
      rulesQuery.isError || (dayMessage === null && slotsQuery.isError)
        ? () => void (rulesQuery.isError ? rulesQuery.refetch() : slotsQuery.refetch())
        : null,
    /**
     * The chosen slot as the API takes it, or null when it no longer suits the
     * items: the clock on screen can be up to a tick old.
     */
    pickupAt(): string | null {
      if (!pickupWindow || pickupProblem(slot, pickupWindow, new Date(), items) !== null) {
        return null;
      }
      return bakeryPickupISOFromLocalInput(slot);
    },
  };
}
