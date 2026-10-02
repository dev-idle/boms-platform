import { shiftDay } from "@/lib/validation/pickup";

/** The Monday (YYYY-MM-DD) starting the bakery week `day` falls in: a week runs Monday to Sunday. */
export function weekStartOf(day: string): string {
  const weekday = new Date(`${day}T00:00:00Z`).getUTCDay();
  return shiftDay(day, -((weekday + 6) % 7));
}
