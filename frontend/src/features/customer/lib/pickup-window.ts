import { formatOrderTypeLabel, type Fulfillment, type OrderType } from "@/lib/schemas/order";
import { clockToMinutes, formatClockMinutes } from "@/lib/validation/clock";
import type { PickupFulfillment, PickupProblem, PickupWindow } from "@/lib/validation/pickup";

import type { PickupRules } from "../schemas";

/** The API's pickup rules in the shape the pickup checks read. */
export function toPickupWindow(rules: PickupRules): PickupWindow {
  return {
    opensAtMinutes: clockToMinutes(rules.opens_at),
    closesAtMinutes: clockToMinutes(rules.closes_at),
    slotMinutes: rules.slot_minutes,
    preorderLeadMinutes: rules.preorder_min_lead_minutes,
    instantPrepMinutes: rules.instant_prep_minutes,
    maxAdvanceDays: rules.max_advance_days,
    closedDates: new Map(rules.closed_dates.map((closed) => [closed.date, closed.reason])),
  };
}

/** What items ask of the bakery, in the shape the pickup checks read. */
export function toPickupFulfillment(fulfillment: Fulfillment): PickupFulfillment {
  return {
    kitchen: fulfillment.has_kitchen_items,
    leadMinutes: fulfillment.lead_minutes,
    soldOutOn: fulfillment.sold_out_on,
  };
}

/** A duration in minutes as people say it, e.g. 90 → "1 hour 30 minutes". */
function formatLead(minutes: number): string {
  const hours = Math.floor(minutes / 60);
  const rest = minutes % 60;
  const parts = [
    hours > 0 ? `${hours} hour${hours === 1 ? "" : "s"}` : null,
    rest > 0 || hours === 0 ? `${rest} minute${rest === 1 ? "" : "s"}` : null,
  ];
  return parts.filter(Boolean).join(" ");
}

function hoursRange(pickupWindow: PickupWindow): string {
  return `${formatClockMinutes(pickupWindow.opensAtMinutes)}–${formatClockMinutes(pickupWindow.closesAtMinutes)}`;
}

/** The notice an order of this type with these items waits. */
function noticeMinutes(type: OrderType, pickupWindow: PickupWindow, items: PickupFulfillment): number {
  const prep = type === "instant" ? pickupWindow.instantPrepMinutes : pickupWindow.preorderLeadMinutes;
  return Math.max(prep, items.leadMinutes);
}

/** The hint under the pickup picker, built from the rules in force. */
export function pickupHint(pickupWindow: PickupWindow): string {
  return (
    `Pickup slots every ${pickupWindow.slotMinutes} minutes, ${hoursRange(pickupWindow)} bakery time, ` +
    `up to ${pickupWindow.maxAdvanceDays} day${pickupWindow.maxAdvanceDays === 1 ? "" : "s"} ahead.`
  );
}

/** How the order will be prepared at the chosen time, and the notice it needs. */
export function orderTypeNote(
  type: OrderType,
  pickupWindow: PickupWindow,
  items: PickupFulfillment,
): string {
  const minutes = noticeMinutes(type, pickupWindow, items);
  const notice = formatLead(minutes);
  const how =
    type === "instant"
      ? minutes === 0
        ? "ready-made items, ready as soon as you order"
        : `ready-made items, packed ${notice} after you order`
      : `place it at least ${notice} before pickup`;
  return `${formatOrderTypeLabel(type)}: ${how}.`;
}

/** What to tell the customer about a pickup time checkout would refuse. */
export function pickupProblemMessage(
  problem: Exclude<PickupProblem, "missing">,
  pickupWindow: PickupWindow,
  items: PickupFulfillment,
  type: OrderType,
  closedReason?: string,
): string {
  switch (problem) {
    case "too_soon":
      return `Pickup must be at least ${formatLead(noticeMinutes(type, pickupWindow, items))} from now.`;
    case "too_far":
      return `Pickup can be at most ${pickupWindow.maxAdvanceDays} day${pickupWindow.maxAdvanceDays === 1 ? "" : "s"} ahead.`;
    case "closed_day":
      return closedReason
        ? `The bakery is closed that day (${closedReason}). Choose another day.`
        : "The bakery is closed that day. Choose another day.";
    case "sold_out":
      return "An item has sold out for that day. Choose a later day.";
    case "outside_hours":
      return `Pickup must be between ${hoursRange(pickupWindow)} bakery time.`;
    case "off_slot":
      return "Choose one of the pickup slots.";
  }
}
