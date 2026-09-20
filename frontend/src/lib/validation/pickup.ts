import { PICKUP_RULES } from "@/constants/pickup";

const MINUTE_MS = 60 * 1000;
const HOUR_MS = 60 * MINUTE_MS;
const DAY_MS = 24 * HOUR_MS;
const BAKERY_OFFSET_MS = PICKUP_RULES.utcOffsetMinutes * MINUTE_MS;

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

function bakeryOffsetSuffix(): string {
  const total = PICKUP_RULES.utcOffsetMinutes;
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
 * Earliest valid pickup slot: now + lead time, snapped forward into business hours.
 * Mirrors backend ValidatePickupAt so the default value always passes checkout.
 */
export function defaultPickupLocalInputValue(now = new Date()): string {
  const wall = toBakeryWallClock(
    new Date(now.getTime() + PICKUP_RULES.minLeadHours * HOUR_MS),
  );
  const hour = wall.getUTCHours();
  if (hour < PICKUP_RULES.openHour) {
    wall.setUTCHours(PICKUP_RULES.openHour, 0, 0, 0);
  } else if (hour >= PICKUP_RULES.closeHour) {
    wall.setUTCDate(wall.getUTCDate() + 1);
    wall.setUTCHours(PICKUP_RULES.openHour, 0, 0, 0);
  }
  return formatWallClock(wall);
}

export function minPickupLocalInputValue(now = new Date()): string {
  return defaultPickupLocalInputValue(now);
}

export function maxPickupLocalInputValue(now = new Date()): string {
  return formatPickupLocalInputValue(
    new Date(now.getTime() + PICKUP_RULES.maxAdvanceDays * DAY_MS),
  );
}

/**
 * Client-side mirror of backend pickup rules (lead time, advance limit, business hours).
 * Returns false for malformed input; the backend remains the authority.
 */
export function isPickupLocalValueValid(
  localValue: string,
  now = new Date(),
): boolean {
  const instant = pickupInstantFromLocalInput(localValue);
  if (!instant) {
    return false;
  }
  const earliest = now.getTime() + PICKUP_RULES.minLeadHours * HOUR_MS;
  const latest = now.getTime() + PICKUP_RULES.maxAdvanceDays * DAY_MS;
  if (instant.getTime() < earliest || instant.getTime() > latest) {
    return false;
  }
  const hour = toBakeryWallClock(instant).getUTCHours();
  return hour >= PICKUP_RULES.openHour && hour < PICKUP_RULES.closeHour;
}

/** Formats a pickup instant in bakery time so all roles see the scheduled wall-clock slot. */
export function formatPickupDateTime(iso: string): string {
  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: PICKUP_RULES.timeZone,
  }).format(new Date(iso));
}

/** Wall-clock pickup time for kitchen queue rows — tabular, no date. */
export function formatPickupWallTime(iso: string): string {
  return new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    minute: "2-digit",
    timeZone: PICKUP_RULES.timeZone,
  }).format(new Date(iso));
}
