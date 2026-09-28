/** The bakery's time zone — fixed, as in backend `domain/store.Location`. */
export const PICKUP_ZONE = {
  timeZone: "Asia/Ho_Chi_Minh",
  /** Fixed offset for Asia/Ho_Chi_Minh (no DST); single source for serialization math. */
  utcOffsetMinutes: 7 * 60,
} as const;

export const PICKUP_COPY = {
  label: "Pickup time",
  required: "Choose a pickup time to continue checkout.",
  loading: "Loading pickup times…",
  unavailable: "Pickup times could not be loaded. Try again in a moment.",
  noOpenTime: "The bakery has no open pickup time in its booking window.",
} as const;
