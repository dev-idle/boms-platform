import { z } from "zod";

/** Order lifecycle — mirrors backend `domain/order.Status`; the single frontend source. */
export const orderStatusSchema = z.enum([
  "pending",
  "confirmed",
  "in_production",
  "ready",
  "cancelled",
  "fulfilled",
]);

export type OrderStatus = z.infer<typeof orderStatusSchema>;
