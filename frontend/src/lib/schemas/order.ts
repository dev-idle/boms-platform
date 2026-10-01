import { z } from "zod";

import { productOptionGroupSchema } from "@/lib/schemas/catalog";
import { apiDateTimeSchema } from "@/lib/validation/datetime";

/** Order lifecycle — mirrors backend `domain/order.Status`; the single frontend source. */
export const orderStatusSchema = z.enum([
  "awaiting_payment",
  "pending",
  "confirmed",
  "in_production",
  "ready",
  "cancelled",
  "fulfilled",
  "expired",
  "no_show",
]);

export type OrderStatus = z.infer<typeof orderStatusSchema>;

/**
 * How a customer configured a custom item: the options chosen with what each
 * added, a message for the cake and a reference photo (empty when none).
 */
export const customizationSchema = z.object({
  options: z.array(
    z.object({
      group: productOptionGroupSchema,
      label: z.string().min(1),
      price_delta_cents: z.number().int().min(0),
    }),
  ),
  message: z.string(),
  reference_image_url: z.union([z.url({ protocol: /^https$/ }), z.literal("")]),
});

export type Customization = z.infer<typeof customizationSchema>;

/** Where an order was taken — mirrors backend `domain/order.Channel`. Staff take counter and phone orders, paid in cash at pickup. */
export const orderChannelSchema = z.enum(["online", "counter", "phone"]);

export type OrderChannel = z.infer<typeof orderChannelSchema>;

/** Where an order was taken, as the counter reads it. */
export function formatOrderChannelLabel(channel: OrderChannel): string {
  switch (channel) {
    case "online":
      return "Online order";
    case "counter":
      return "Counter order";
    case "phone":
      return "Phone order";
  }
}

/** Where an order's payment stands — null on the order before the customer starts paying. */
export const orderPaymentSchema = z.object({
  provider: z.enum(["paypal", "cash"]),
  status: z.enum(["created", "pending", "captured", "denied", "refunded"]),
  captured_at: apiDateTimeSchema.nullable(),
  /** Set once the order is cancelled; `refunded_at` once the money went back. */
  refund_requested_at: apiDateTimeSchema.nullable(),
  refunded_at: apiDateTimeSchema.nullable(),
});

export type OrderPayment = z.infer<typeof orderPaymentSchema>;

/** What items ask of the bakery: whether the kitchen makes any, the longest notice one needs, and the day one ran out on. */
export const fulfillmentSchema = z.object({
  has_kitchen_items: z.boolean(),
  lead_minutes: z.number().int().min(0),
  /** The bakery day one of the items ran out on, when they cannot be collected; null when none has. */
  sold_out_on: z.iso.date().nullable(),
});

export type Fulfillment = z.infer<typeof fulfillmentSchema>;

/** How an order is prepared — mirrors backend `domain/order.Type`. */
export const orderTypeSchema = z.enum(["instant", "pre_order"]);

export type OrderType = z.infer<typeof orderTypeSchema>;

/** The order type as the customer and the counter read it. */
export function formatOrderTypeLabel(type: OrderType): string {
  return type === "instant" ? "Instant pickup" : "Pre-order";
}

/** CH-YYMMDD-NNN: the bakery day an order was placed and its number that day (backend `order.Code`). */
export const orderCodeSchema = z.string().regex(/^CH-\d{6}-\d{3,}$/);

/** A status an order entered and when — one entry of `timeline` on an order detail. */
export const orderTimelineEntrySchema = z.object({
  status: orderStatusSchema,
  /** Why the bakery cancelled the order; null for any other move. */
  reason: z.string().nullable(),
  at: apiDateTimeSchema,
});

export type OrderTimelineEntry = z.infer<typeof orderTimelineEntrySchema>;

/**
 * Whether this status is the one being applied right now. A queue offers several
 * transitions at once, so only the button that was pressed may say it is working
 * — the rest are simply held.
 */
export function isApplyingOrderStatus(
  mutation: { isPending: boolean; variables?: { status: OrderStatus } },
  status: OrderStatus,
): boolean {
  return mutation.isPending && mutation.variables?.status === status;
}
