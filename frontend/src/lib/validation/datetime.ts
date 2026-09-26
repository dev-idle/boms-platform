import { z } from "zod";

/**
 * RFC 3339 timestamps from the Go API (Fiber JSON) include a numeric offset
 * (e.g. `+07:00`). Zod's default `datetime()` only accepts UTC `Z` suffixes.
 */
export const apiDateTimeSchema = z.string().datetime({ offset: true });

/**
 * The near end of a default promotion window. A combo or code a manager is
 * creating right now starts right now — anything later is a delay they did not
 * ask for and may not notice, and the backend only asks that the window ends
 * after it begins.
 */
export function windowStartsNow(): string {
  return new Date().toISOString();
}

/** The far end of that window, `days` out. */
export function windowEndsInDays(days: number): string {
  return new Date(Date.now() + days * 24 * 60 * 60 * 1000).toISOString();
}

export function formatDateTime(iso: string): string {
  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(iso));
}

/**
 * The same timestamp split in two, for a table cell: the date is what a column
 * is scanned by, the time is the detail under it. One line would need a column
 * half again as wide, which is taken from the name beside it.
 */
export function splitDateTime(iso: string): { date: string; time: string } {
  const value = new Date(iso);

  return {
    date: new Intl.DateTimeFormat("en-US", { dateStyle: "medium" }).format(value),
    time: new Intl.DateTimeFormat("en-US", { timeStyle: "short" }).format(value),
  };
}
