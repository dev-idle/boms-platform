import { z } from "zod";

import { apiDateTimeSchema } from "@/lib/validation/datetime";

/** What went wrong with an order — mirrors backend `order.IncidentType`. */
export const incidentTypeSchema = z.enum([
  "bakery_cancelled",
  "ready_late",
  "no_show",
  "payment_failed",
  "payment_expired",
  "refunded",
  "payment_anomaly",
  "wrong_items",
  "custom_mismatch",
  "other",
]);

export type IncidentType = z.infer<typeof incidentTypeSchema>;

/** The types staff report; the system records every other as it happens. */
export const REPORTED_INCIDENT_TYPES = ["wrong_items", "custom_mismatch", "other"] as const satisfies readonly IncidentType[];

export const INCIDENT_TYPE_LABEL: Record<IncidentType, string> = {
  bakery_cancelled: "Cancelled by the bakery",
  ready_late: "Ready after pickup time",
  no_show: "Not collected",
  payment_failed: "Payment failed",
  payment_expired: "Not paid in time",
  refunded: "Refunded",
  payment_anomaly: "Unusual payment activity",
  wrong_items: "Wrong items",
  custom_mismatch: "Custom cake not as asked",
  other: "Other problem",
};

/**
 * An incident recorded with an order: `source` is `auto` when the system
 * recorded it, `manual` when staff reported it; `note` is what staff wrote, or
 * the reason the bakery cancelled.
 */
export const orderIncidentSchema = z.object({
  id: z.uuid(),
  type: incidentTypeSchema,
  source: z.enum(["auto", "manual"]),
  note: z.string().min(1).nullable(),
  created_at: apiDateTimeSchema,
});

export type OrderIncident = z.infer<typeof orderIncidentSchema>;
