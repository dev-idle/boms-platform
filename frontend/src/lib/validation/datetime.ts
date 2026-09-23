import { z } from "zod";

/**
 * RFC 3339 timestamps from the Go API (Fiber JSON) include a numeric offset
 * (e.g. `+07:00`). Zod's default `datetime()` only accepts UTC `Z` suffixes.
 */
export const apiDateTimeSchema = z.string().datetime({ offset: true });

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
