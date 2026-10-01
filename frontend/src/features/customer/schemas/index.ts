import { z } from "zod";

import { catalogSlugSchema } from "@/lib/validation/catalog";
import { CAKE_MESSAGE_MAX } from "@/lib/schemas/catalog";
import {
  customizationSchema,
  orderCodeSchema,
  orderPaymentSchema,
  orderStatusSchema,
  orderTimelineEntrySchema,
  orderTypeSchema,
} from "@/lib/schemas/order";
import { ticketSummarySchema } from "@/lib/schemas/ticket";
import { clockTimeSchema } from "@/lib/validation/clock";
import { apiDateTimeSchema } from "@/lib/validation/datetime";

const cartItemSchema = z.object({
  id: z.uuid(),
  line_type: z.enum(["product", "combo"]),
  product_id: z.uuid().nullable().optional(),
  combo_id: z.uuid().nullable().optional(),
  name: z.string(),
  slug: catalogSlugSchema,
  quantity: z.number().int().min(1).max(99),
  unit_price_cents: z.number().int().min(0),
  line_total_cents: z.number().int().min(0),
  is_available: z.boolean(),
  /** How the line is configured, null for a plain line. */
  customization: customizationSchema.nullable(),
});

const cartDiscountSchema = z.object({
  code: z.string().min(1),
  discount_type: z.enum(["percent", "fixed_cents"]),
  value: z.number().int().min(1),
  discount_cents: z.number().int().min(0),
});

/** What items ask of the bakery: whether the kitchen makes any, and the longest notice one needs. */
const fulfillmentSchema = z.object({
  has_kitchen_items: z.boolean(),
  lead_minutes: z.number().int().min(0),
});

export const cartSchema = z.object({
  id: z.uuid(),
  items: z.array(cartItemSchema),
  subtotal_cents: z.number().int().min(0),
  discount: cartDiscountSchema.nullable().optional(),
  discount_cents: z.number().int().min(0),
  total_cents: z.number().int().min(0),
  checkout_ready: z.boolean(),
  /** What the items ask of the bakery; the pickup picker offers times to suit. */
  fulfillment: fulfillmentSchema,
});

/**
 * How a customer configures a customizable product: one option from each
 * group it offers, a message for the cake, and a photo they uploaded with their
 * word that they may share it.
 */
export const cartCustomizationInputSchema = z
  .object({
    option_ids: z.array(z.uuid("Choose an option")).max(3),
    message: z
      .string()
      .trim()
      .max(CAKE_MESSAGE_MAX, `At most ${CAKE_MESSAGE_MAX} characters`)
      .refine((value) => !/[\p{Cc}\p{Cf}]/u.test(value), "Use plain text only"),
    reference_image_url: z.union([z.url({ protocol: /^https$/ }), z.literal("")]),
    reference_rights_confirmed: z.boolean(),
  })
  .refine((value) => value.reference_image_url === "" || value.reference_rights_confirmed, {
    error: "Confirm you may share this photo",
    path: ["reference_rights_confirmed"],
  });

export const addCartItemInputSchema = z
  .object({
    product_id: z.uuid().optional(),
    combo_id: z.uuid().optional(),
    quantity: z.number().int().min(1).max(99).default(1),
    customization: cartCustomizationInputSchema.optional(),
  })
  .refine(
    (value) =>
      (value.product_id !== undefined) !== (value.combo_id !== undefined),
    { error: "Provide exactly one of product_id or combo_id" },
  );

export const updateCartItemInputSchema = z.object({
  quantity: z.number().int().min(1).max(99),
});

export const applyCartDiscountInputSchema = z.object({
  code: z.string().min(3).max(64),
});

export const checkoutInputSchema = z.object({
  pickup_at: apiDateTimeSchema,
});

/** `PATCH /orders/:id/pickup` — another slot for an order not being made yet. */
export const rescheduleInputSchema = z.object({
  pickup_at: apiDateTimeSchema,
});

const orderItemSchema = z.object({
  id: z.uuid(),
  line_type: z.enum(["product", "combo"]),
  product_id: z.uuid().nullable().optional(),
  combo_id: z.uuid().nullable().optional(),
  name: z.string(),
  slug: catalogSlugSchema,
  quantity: z.number().int().min(1),
  unit_price_cents: z.number().int().min(0),
  line_total_cents: z.number().int().min(0),
  /** How the customer configured it, null for a plain item. */
  customization: customizationSchema.nullable(),
});

export const orderSchema = z.object({
  id: z.uuid(),
  code: orderCodeSchema,
  status: orderStatusSchema,
  order_type: orderTypeSchema,
  subtotal_cents: z.number().int().min(0),
  discount_cents: z.number().int().min(0),
  total_cents: z.number().int().min(0),
  discount_code_snapshot: z.string().nullable().optional(),
  pickup_at: apiDateTimeSchema.nullable().optional(),
  items: z.array(orderItemSchema),
  timeline: z.array(orderTimelineEntrySchema),
  tickets: z.array(ticketSummarySchema),
  payment: orderPaymentSchema.nullable(),
  /** What the items ask of the bakery, for moving the pickup; null once the order is being made. */
  fulfillment: fulfillmentSchema.nullable(),
  /** When the order expires if still unpaid; null on an order placed before online payment. */
  payment_due_at: apiDateTimeSchema.nullable(),
  /** What the customer gives at the counter, from when the bakery accepts the order until it is collected. */
  pickup_code: z.string().regex(/^\d{4}$/).nullable(),
  created_at: apiDateTimeSchema,
  updated_at: apiDateTimeSchema,
});

/** POST /orders/:id/payment — the PayPal page the customer approves the payment on. */
export const paymentStartSchema = z.object({
  approve_url: z.url({ protocol: /^https$/, hostname: /(^|\.)paypal\.com$/ }),
});

/** POST /orders/:id/payment/capture — taken, or held by PayPal for review. */
export const paymentCaptureSchema = z.object({
  status: z.enum(["captured", "pending"]),
});

export const orderSummarySchema = z.object({
  id: z.uuid(),
  code: orderCodeSchema,
  status: orderStatusSchema,
  total_cents: z.number().int().min(0),
  item_count: z.number().int().min(0),
  pickup_at: apiDateTimeSchema.nullable().optional(),
  created_at: apiDateTimeSchema,
});

/** GET /api/v1/store/pickup-rules — the window checkout holds a pickup to. */
export const pickupRulesSchema = z.object({
  opens_at: clockTimeSchema,
  closes_at: clockTimeSchema,
  slot_minutes: z.number().int().min(1),
  preorder_min_lead_minutes: z.number().int().min(0),
  instant_prep_minutes: z.number().int().min(0),
  max_advance_days: z.number().int().min(1),
  closed_dates: z.array(
    z.object({
      date: z.iso.date(),
      reason: z.string().min(1),
    }),
  ),
});

/** GET /api/v1/store/pickup-slots?date= — one day's slots, full ones marked. */
export const pickupSlotsSchema = z.object({
  date: z.iso.date(),
  slots: z.array(
    z.object({
      starts_at: apiDateTimeSchema,
      full: z.boolean(),
    }),
  ),
});

/** Order history: a page, narrowed by status and by the bakery days orders were placed on. */
export const ordersListFilterSchema = z
  .object({
    page: z.number().int().min(1).default(1),
    page_size: z.number().int().min(1).max(100).default(20),
    status: orderStatusSchema.optional(),
    from: z.iso.date().optional(),
    to: z.iso.date().optional(),
  })
  .refine((filter) => !filter.from || !filter.to || filter.from <= filter.to, {
    path: ["to"],
    message: "The last day cannot be before the first",
  });

export type Cart = z.infer<typeof cartSchema>;
export type CartItem = z.infer<typeof cartItemSchema>;
export type AddCartItemInput = z.infer<typeof addCartItemInputSchema>;
export type CartCustomizationInput = z.input<typeof cartCustomizationInputSchema>;
export type UpdateCartItemInput = z.infer<typeof updateCartItemInputSchema>;
export type ApplyCartDiscountInput = z.infer<typeof applyCartDiscountInputSchema>;
export type CheckoutInput = z.infer<typeof checkoutInputSchema>;
export type RescheduleInput = z.infer<typeof rescheduleInputSchema>;
export type Fulfillment = z.infer<typeof fulfillmentSchema>;
export type Order = z.infer<typeof orderSchema>;
export type PaymentStart = z.infer<typeof paymentStartSchema>;
export type PaymentCapture = z.infer<typeof paymentCaptureSchema>;
type OrderSummary = z.infer<typeof orderSummarySchema>;
export type OrdersListFilterInput = z.infer<typeof ordersListFilterSchema>;
export type PickupRules = z.infer<typeof pickupRulesSchema>;
export type PickupSlots = z.infer<typeof pickupSlotsSchema>;

export type OrdersListResult = {
  orders: OrderSummary[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
};
