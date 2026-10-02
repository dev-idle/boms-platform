import type { OrderStatus } from "@/lib/schemas/order";

import type { StaffPickup } from "../schemas";

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

/** An order nobody has collected though its slot, `slotMinutes` long, has ended. */
export function isLatePickup(pickup: StaffPickup, slotMinutes: number, now: Date): boolean {
  return (
    STILL_TO_COLLECT.has(pickup.status) &&
    Date.parse(pickup.pickup_at) + slotMinutes * 60_000 <= now.getTime()
  );
}
