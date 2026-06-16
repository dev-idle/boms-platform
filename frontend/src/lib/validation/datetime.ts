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
