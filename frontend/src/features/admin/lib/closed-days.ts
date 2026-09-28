import { formatPickupLocalInputValue } from "@/lib/validation/pickup";

const DAY_MS = 24 * 60 * 60 * 1000;
/** How far ahead a closure may be planned — backend `store.ClosedDateWindow`. */
const CLOSED_DAY_HORIZON_DAYS = 365;

/** The days a closure may be added for: bakery-local today through a year ahead. */
export function closedDayBounds(now = new Date()): { min: string; max: string } {
  const today = formatPickupLocalInputValue(now).slice(0, 10);
  const last = new Date(Date.parse(`${today}T00:00:00Z`) + CLOSED_DAY_HORIZON_DAYS * DAY_MS);
  return { min: today, max: last.toISOString().slice(0, 10) };
}

/** A calendar day (YYYY-MM-DD) as people read it, e.g. "Tue, Oct 20, 2026". */
export function formatClosedDay(day: string): string {
  return new Intl.DateTimeFormat("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
    year: "numeric",
    timeZone: "UTC",
  }).format(new Date(`${day}T00:00:00Z`));
}
