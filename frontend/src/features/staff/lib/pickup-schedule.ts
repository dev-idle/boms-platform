import type { Fulfillment, OrderStatus } from "@/lib/schemas/order";

import type { StaffPickup } from "../schemas";

/**
 * Items that ask nothing of the kitchen, so only the counter's own notice
 * applies: an order not priced yet, or a pickup staff move at their own call.
 */
export const NO_ITEM_NEEDS: Fulfillment = { has_kitchen_items: false, lead_minutes: 0, sold_out_on: null };

/** The pickups booked for one time on the day's schedule. */
type PickupSlot = {
  at: string;
  pickups: StaffPickup[];
};

const STILL_TO_COLLECT: ReadonlySet<OrderStatus> = new Set([
  "pending",
  "confirmed",
  "in_production",
  "ready",
]);

/** Groups a day's pickups, sorted by time as the API returns them, into their slots. */
export function groupPickupsBySlot(pickups: ReadonlyArray<StaffPickup>): PickupSlot[] {
  const slots: PickupSlot[] = [];
  for (const pickup of pickups) {
    const last = slots[slots.length - 1];
    if (last && Date.parse(last.at) === Date.parse(pickup.pickup_at)) {
      last.pickups.push(pickup);
    } else {
      slots.push({ at: pickup.pickup_at, pickups: [pickup] });
    }
  }
  return slots;
}

/** Why the counter's part of this order cannot be marked ready, or undefined when it can. */
export function counterReadyBlockedReason(pickup: StaffPickup): string | undefined {
  const ticket = pickup.counter_ticket;
  if (!ticket) {
    return "The kitchen makes this order";
  }
  if (ticket.status === "ready") {
    return "Counter part ready";
  }
  if (pickup.status === "pending") {
    return "Accept the order first";
  }
  if (pickup.status !== "confirmed" && pickup.status !== "in_production") {
    return "Nothing left to prepare";
  }
  return undefined;
}

/** An order nobody has collected though its slot, `slotMinutes` long, has ended. */
export function isLatePickup(pickup: StaffPickup, slotMinutes: number, now: Date): boolean {
  return (
    STILL_TO_COLLECT.has(pickup.status) &&
    Date.parse(pickup.pickup_at) + slotMinutes * 60_000 <= now.getTime()
  );
}
