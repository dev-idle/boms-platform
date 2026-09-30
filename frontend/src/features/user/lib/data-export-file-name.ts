import { BRAND } from "@/constants/brand";
import { bakeryDayOf } from "@/lib/validation/pickup";

/** "choux-my-data-2026-09-29.json": whose data it is and the bakery day it was taken. */
export function dataExportFileName(exportedAt: Date): string {
  return `${BRAND.name.toLowerCase()}-my-data-${bakeryDayOf(exportedAt)}.json`;
}
