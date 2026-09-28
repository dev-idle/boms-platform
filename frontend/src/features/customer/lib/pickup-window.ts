import { clockToMinutes, formatClockMinutes } from "@/lib/validation/clock";
import type { PickupProblem, PickupWindow } from "@/lib/validation/pickup";

import type { PickupRules } from "../schemas";

/** The API's pickup rules in the shape the pickup checks read. */
export function toPickupWindow(rules: PickupRules): PickupWindow {
  return {
    opensAtMinutes: clockToMinutes(rules.opens_at),
    closesAtMinutes: clockToMinutes(rules.closes_at),
    minLeadMinutes: rules.preorder_min_lead_minutes,
    maxAdvanceDays: rules.max_advance_days,
    closedDates: new Map(rules.closed_dates.map((closed) => [closed.date, closed.reason])),
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

/** The hint under the pickup picker, built from the rules in force. */
export function pickupHint(pickupWindow: PickupWindow): string {
  return (
    `Collect your order at the bakery between ${hoursRange(pickupWindow)} bakery time, ` +
    `at least ${formatLead(pickupWindow.minLeadMinutes)} from now and up to ` +
    `${pickupWindow.maxAdvanceDays} day${pickupWindow.maxAdvanceDays === 1 ? "" : "s"} ahead.`
  );
}

/** What to tell the customer about a pickup time checkout would refuse. */
export function pickupProblemMessage(
  problem: Exclude<PickupProblem, "missing">,
  pickupWindow: PickupWindow,
  closedReason?: string,
): string {
  switch (problem) {
    case "too_soon":
      return `Pickup must be at least ${formatLead(pickupWindow.minLeadMinutes)} from now.`;
    case "too_far":
      return `Pickup can be at most ${pickupWindow.maxAdvanceDays} day${pickupWindow.maxAdvanceDays === 1 ? "" : "s"} ahead.`;
    case "closed_day":
      return closedReason
        ? `The bakery is closed that day (${closedReason}). Choose another day.`
        : "The bakery is closed that day. Choose another day.";
    case "outside_hours":
      return `Pickup must be between ${hoursRange(pickupWindow)} bakery time.`;
  }
}
