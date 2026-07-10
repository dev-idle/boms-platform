/** Bakery pickup window — keep aligned with backend domain/order/pickup.go */
export const PICKUP_RULES = {
  minLeadHours: 2,
  maxAdvanceDays: 14,
  openHour: 8,
  closeHour: 18,
  timeZone: "Asia/Ho_Chi_Minh",
  /** Fixed offset for Asia/Ho_Chi_Minh (no DST); single source for serialization math. */
  utcOffsetMinutes: 7 * 60,
} as const;

export const PICKUP_COPY = {
  label: "Pickup time",
  hint: "Select when you will collect your order at the bakery (8:00 AM–6:00 PM, at least 2 hours from now, up to 14 days ahead).",
  required: "Choose a pickup time to continue checkout.",
  outsideWindow:
    "Pickup must be 8:00 AM–6:00 PM bakery time, at least 2 hours from now and within 14 days.",
} as const;
