import { z } from "zod";

/** HH:MM wall-clock time in bakery time, 00:00 to 23:59. */
export const clockTimeSchema = z
  .string()
  .regex(/^(?:[01]\d|2[0-3]):[0-5]\d$/, "Use HH:MM, e.g. 08:30");

/** Minutes after midnight for a valid HH:MM time. */
export function clockToMinutes(clock: string): number {
  const [hours, minutes] = clock.split(":").map(Number);
  return hours * 60 + minutes;
}

/** Minutes after midnight as a readable time of day, e.g. 510 → "8:30 AM". */
export function formatClockMinutes(minutes: number): string {
  return new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    minute: "2-digit",
    timeZone: "UTC",
  }).format(new Date(Date.UTC(1970, 0, 1, 0, minutes)));
}
