/** The bakery's time zone — fixed, as in backend `domain/store.Location`. */
export const PICKUP_ZONE = {
  timeZone: "Asia/Ho_Chi_Minh",
  /** Fixed offset for Asia/Ho_Chi_Minh (no DST); single source for serialization math. */
  utcOffsetMinutes: 7 * 60,
} as const;

export const PICKUP_COPY = {
  dayLabel: "Pickup day",
  label: "Pickup time",
  full: "Full",
  tooSoon: "Too soon",
  slotPlaceholder: "Choose a time",
  required: "Choose a pickup day and time to continue checkout.",
  loading: "Loading pickup times…",
  unavailable: "Pickup times could not be loaded. Try again in a moment.",
  slotsUnavailable: "The pickup slots for that day could not be loaded. Try again in a moment.",
  noOpenTime: "The bakery has no open pickup time in its booking window.",
  noSlotLeft: "No pickup time is left on that day. Choose another day.",
  slotFull: "That pickup time has just filled up. Choose another time.",
} as const;
