import { PICKUP_ZONE } from "@/constants/pickup";

const MINUTE_MS = 60 * 1000;
const DAY_MS = 24 * 60 * MINUTE_MS;
/** Suggested pickups fall on five-minute marks, so the default is a time people say. */
const SUGGESTED_STEP_MS = 5 * MINUTE_MS;
const BAKERY_OFFSET_MS = PICKUP_ZONE.utcOffsetMinutes * MINUTE_MS;

/**
 * The pickup window the bakery's settings allow (GET /store/pickup-rules),
 * in bakery time. Mirrors backend `order.PickupPolicy`.
 */
export type PickupWindow = {
  /** Minutes after bakery midnight when pickups start. */
  opensAtMinutes: number;
  /** Minutes after bakery midnight when pickups stop (exclusive). */
  closesAtMinutes: number;
  minLeadMinutes: number;
  maxAdvanceDays: number;
  /** Closed bakery days (YYYY-MM-DD) and the reason customers see. */
  closedDates: ReadonlyMap<string, string>;
};

/** Why a pickup time cannot be used — the backend's pickup error codes, plus "missing". */
export type PickupProblem =
  | "missing"
  | "too_soon"
  | "too_far"
  | "closed_day"
  | "outside_hours";

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

function wallDay(wall: Date): string {
  return formatWallClock(wall).slice(0, 10);
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

/**
 * What is wrong with a pickup time, or null when checkout would accept it.
 * Checked in the backend's order, so both sides name the same problem.
 */
export function pickupProblem(
  localValue: string,
  pickupWindow: PickupWindow,
  now: Date,
): PickupProblem | null {
  const instant = pickupInstantFromLocalInput(localValue);
  if (!instant) {
    return "missing";
  }
  if (instant.getTime() < now.getTime() + pickupWindow.minLeadMinutes * MINUTE_MS) {
    return "too_soon";
  }
  if (instant.getTime() > now.getTime() + pickupWindow.maxAdvanceDays * DAY_MS) {
    return "too_far";
  }
  const wall = toBakeryWallClock(instant);
  if (pickupWindow.closedDates.has(wallDay(wall))) {
    return "closed_day";
  }
  const minutes = wallMinutes(wall);
  if (minutes < pickupWindow.opensAtMinutes || minutes >= pickupWindow.closesAtMinutes) {
    return "outside_hours";
  }
  return null;
}

/**
 * Earliest pickup checkout accepts: now plus the lead time, rounded up to a
 * five-minute mark, moved forward into opening hours and past closed days.
 * Empty when the booking window holds no open time.
 */
export function defaultPickupLocalInputValue(
  pickupWindow: PickupWindow,
  now: Date,
): string {
  const latest = now.getTime() + pickupWindow.maxAdvanceDays * DAY_MS;
  const earliest = now.getTime() + pickupWindow.minLeadMinutes * MINUTE_MS;
  // The bakery's offset is whole hours, so marks on the epoch are marks on its clock.
  const wall = toBakeryWallClock(
    new Date(Math.ceil(earliest / SUGGESTED_STEP_MS) * SUGGESTED_STEP_MS),
  );
  while (wall.getTime() - BAKERY_OFFSET_MS <= latest) {
    const minutes = wallMinutes(wall);
    if (pickupWindow.closedDates.has(wallDay(wall)) || minutes >= pickupWindow.closesAtMinutes) {
      wall.setUTCDate(wall.getUTCDate() + 1);
      wall.setUTCHours(0, pickupWindow.opensAtMinutes, 0, 0);
      continue;
    }
    if (minutes < pickupWindow.opensAtMinutes) {
      // Opening may lie past the window; the loop checks again.
      wall.setUTCHours(0, pickupWindow.opensAtMinutes, 0, 0);
      continue;
    }
    return formatWallClock(wall);
  }
  return "";
}

/** Latest pickup the booking window allows, as a `datetime-local` bound. */
export function maxPickupLocalInputValue(
  pickupWindow: PickupWindow,
  now: Date,
): string {
  return formatPickupLocalInputValue(
    new Date(now.getTime() + pickupWindow.maxAdvanceDays * DAY_MS),
  );
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
