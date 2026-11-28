"use client";

import { usePickupRules } from "../hooks";
import { hoursRange, toPickupWindow } from "../lib/pickup-window";

/** The opening hours an admin set, e.g. "8:00 AM–6:00 PM"; nothing until they load. */
export function StoreHours() {
  const rules = usePickupRules();
  return rules.data ? hoursRange(toPickupWindow(rules.data)) : null;
}
