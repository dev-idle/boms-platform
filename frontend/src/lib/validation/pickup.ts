import { PICKUP_ZONE } from "@/constants/pickup";
import type { OrderType } from "@/lib/schemas/order";

const MINUTE_MS = 60 * 1000;
const DAY_MS = 24 * 60 * MINUTE_MS;
const BAKERY_OFFSET_MS = PICKUP_ZONE.utcOffsetMinutes * MINUTE_MS;

/**
 * The pickup window the bakery's settings allow (GET /store/pickup-rules),
 * in bakery time. Mirrors backend `order.PickupPolicy`.
 */
export type PickupWindow = {
  /** Minutes after bakery midnight when the first slot starts. */
  opensAtMinutes: number;
  /** Minutes after bakery midnight when pickups stop (exclusive). */
  closesAtMinutes: number;
  /** Slots start every this many minutes from opening. */
  slotMinutes: number;
  /** Notice a pre-order needs. */
  preorderLeadMinutes: number;
  /** Notice the counter needs to pack a same-day order of ready-made items. */
  instantPrepMinutes: number;
  maxAdvanceDays: number;
  /** Closed bakery days (YYYY-MM-DD) and the reason customers see. */
  closedDates: ReadonlyMap<string, string>;
};

/** What a cart's items ask of the bakery — backend `order.Fulfillment`. */
export type PickupFulfillment = {
  /** Any item is made in the kitchen. */
  kitchen: boolean;
  /** The longest notice any item needs. */
  leadMinutes: number;
  /** The bakery day (YYYY-MM-DD) one of the items ran out on, when none may be collected; null if none has. */
  soldOutOn: string | null;
};

/** Why a pickup time cannot be used — the backend's pickup error codes, plus "missing". */
export type PickupProblem =
  | "missing"
  | "too_soon"
  | "too_far"
  | "closed_day"
  | "sold_out"
  | "outside_hours"
  | "off_slot";

function pad2(value: number): string {
  return String(value).padStart(2, "0");
}

/**
 * Shifted Date whose UTC getters/setters read and write bakery wall-clock time.
 * Valid because Asia/Ho_Chi_Minh is a fixed-offset zone (no DST).
 */
function toBakeryWallClock(instant: Date): Date {
  return new Date(instant.getTime() + BAKERY_OFFSET_MS);
}

function formatWallClock(wall: Date): string {
  return (
    `${wall.getUTCFullYear()}-${pad2(wall.getUTCMonth() + 1)}-${pad2(wall.getUTCDate())}` +
    `T${pad2(wall.getUTCHours())}:${pad2(wall.getUTCMinutes())}`
  );
}

function wallMinutes(wall: Date): number {
  return wall.getUTCHours() * 60 + wall.getUTCMinutes();
}

function bakeryOffsetSuffix(): string {
  const total = PICKUP_ZONE.utcOffsetMinutes;
  const sign = total < 0 ? "-" : "+";
  const abs = Math.abs(total);
  return `${sign}${pad2(Math.floor(abs / 60))}:${pad2(abs % 60)}`;
}

/** Formats an instant as a bakery-local `datetime-local` input value. */
export function formatPickupLocalInputValue(date: Date): string {
  return formatWallClock(toBakeryWallClock(date));
}

/** The bakery day (YYYY-MM-DD) an instant falls on. */
export function bakeryDayOf(date: Date): string {
  return formatPickupLocalInputValue(date).slice(0, 10);
}

/** The day (YYYY-MM-DD) `days` after `day`. */
export function shiftDay(day: string, days: number): string {
  const date = new Date(`${day}T00:00:00Z`);
  date.setUTCDate(date.getUTCDate() + days);
  return date.toISOString().slice(0, 10);
}

/** Serializes a bakery-local `datetime-local` value to an RFC3339 instant. */
export function bakeryPickupISOFromLocalInput(localValue: string): string {
  return `${localValue}:00${bakeryOffsetSuffix()}`;
}

/** Parses a bakery-local `datetime-local` value back to a UTC instant, or null if malformed. */
export function pickupInstantFromLocalInput(localValue: string): Date | null {
  if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/.test(localValue)) {
    return null;
  }
  const instant = new Date(bakeryPickupISOFromLocalInput(localValue));
  return Number.isNaN(instant.getTime()) ? null : instant;
}

function orderTypeAt(instant: Date, now: Date, items: PickupFulfillment): OrderType {
  return !items.kitchen && bakeryDayOf(instant) === bakeryDayOf(now) ? "instant" : "pre_order";
}

/**
 * How an order with these items is prepared if collected then: counter items
 * only, collected the day they are ordered, make an instant order; anything
 * else is a pre-order. Null for a value that is not a time.
 */
export function pickupOrderType(
  localValue: string,
  now: Date,
  items: PickupFulfillment,
): OrderType | null {
  const instant = pickupInstantFromLocalInput(localValue);
  return instant ? orderTypeAt(instant, now, items) : null;
}

/**
 * What is wrong with a pickup time for these items, or null when checkout
 * would accept it (bar a full slot, which only the API knows). Checked in the
 * backend's order, so both sides name the same problem.
 */
export function pickupProblem(
  localValue: string,
  pickupWindow: PickupWindow,
  now: Date,
  items: PickupFulfillment,
): PickupProblem | null {
  const instant = pickupInstantFromLocalInput(localValue);
  if (!instant) {
    return "missing";
  }
  const prep =
    orderTypeAt(instant, now, items) === "instant"
      ? pickupWindow.instantPrepMinutes
      : pickupWindow.preorderLeadMinutes;
  if (instant.getTime() < now.getTime() + Math.max(prep, items.leadMinutes) * MINUTE_MS) {
    return "too_soon";
  }
  if (instant.getTime() > now.getTime() + pickupWindow.maxAdvanceDays * DAY_MS) {
    return "too_far";
  }
  const wall = toBakeryWallClock(instant);
  const day = formatWallClock(wall).slice(0, 10);
  if (pickupWindow.closedDates.has(day)) {
    return "closed_day";
  }
  if (day === items.soldOutOn) {
    return "sold_out";
  }
  const minutes = wallMinutes(wall);
  if (minutes < pickupWindow.opensAtMinutes || minutes >= pickupWindow.closesAtMinutes) {
    return "outside_hours";
  }
  if ((minutes - pickupWindow.opensAtMinutes) % pickupWindow.slotMinutes !== 0) {
    return "off_slot";
  }
  return null;
}

/** The slot starts of a bakery day (YYYY-MM-DD), as `datetime-local` values. */
export function pickupSlotsOfDay(day: string, pickupWindow: PickupWindow): string[] {
  const slots: string[] = [];
  for (
    let minutes = pickupWindow.opensAtMinutes;
    minutes < pickupWindow.closesAtMinutes;
    minutes += pickupWindow.slotMinutes
  ) {
    slots.push(`${day}T${pad2(Math.floor(minutes / 60))}:${pad2(minutes % 60)}`);
  }
  return slots;
}

/** The last bakery day (YYYY-MM-DD) the booking window reaches. */
export function lastPickupDay(pickupWindow: PickupWindow, now: Date): string {
  return bakeryDayOf(new Date(now.getTime() + pickupWindow.maxAdvanceDays * DAY_MS));
}

/**
 * The first slot checkout would accept for these items, as a `datetime-local`
 * value, whether or not it has room — only the API knows that. Empty when the
 * booking window holds none.
 */
export function earliestPickupLocalValue(
  pickupWindow: PickupWindow,
  now: Date,
  items: PickupFulfillment,
): string {
  for (let offset = 0; offset <= pickupWindow.maxAdvanceDays; offset += 1) {
    const day = bakeryDayOf(new Date(now.getTime() + offset * DAY_MS));
    if (pickupWindow.closedDates.has(day)) {
      continue;
    }
    const open = pickupSlotsOfDay(day, pickupWindow).find(
      (slot) => pickupProblem(slot, pickupWindow, now, items) === null,
    );
    if (open) {
      return open;
    }
  }
  return "";
}

/** Formats a pickup instant in bakery time so all roles see the scheduled wall-clock slot. */
export function formatPickupDateTime(iso: string): string {
  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: PICKUP_ZONE.timeZone,
  }).format(new Date(iso));
}

/** Wall-clock pickup time for kitchen queue rows — tabular, no date. */
export function formatPickupWallTime(iso: string): string {
  return new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    minute: "2-digit",
    timeZone: PICKUP_ZONE.timeZone,
  }).format(new Date(iso));
}
