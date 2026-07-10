import { z } from "zod";

import { catalogSlugSchema } from "@/lib/validation/catalog";
import { apiDateTimeSchema } from "@/lib/validation/datetime";

export const cartItemSchema = z.object({
  id: z.string().uuid(),
  line_type: z.enum(["product", "combo"]),
  product_id: z.string().uuid().nullable().optional(),
  combo_id: z.string().uuid().nullable().optional(),
  name: z.string(),
  slug: catalogSlugSchema,
  quantity: z.number().int().min(1).max(99),
  unit_price_cents: z.number().int().min(0),
  line_total_cents: z.number().int().min(0),
  is_available: z.boolean(),
});

export const cartDiscountSchema = z.object({
  code: z.string().min(1),
  discount_type: z.enum(["percent", "fixed_cents"]),
  value: z.number().int().min(1),
  discount_cents: z.number().int().min(0),
});

export const cartSchema = z.object({
  id: z.string().uuid(),
  items: z.array(cartItemSchema),
  subtotal_cents: z.number().int().min(0),
  discount: cartDiscountSchema.nullable().optional(),
  discount_cents: z.number().int().min(0),
  total_cents: z.number().int().min(0),
  checkout_ready: z.boolean(),
});

export const addCartItemInputSchema = z
  .object({
    product_id: z.string().uuid().optional(),
    combo_id: z.string().uuid().optional(),
    quantity: z.number().int().min(1).max(99).default(1),
  })
  .refine(
    (value) =>
      (value.product_id !== undefined) !== (value.combo_id !== undefined),
    { message: "Provide exactly one of product_id or combo_id" },
  );

export const updateCartItemInputSchema = z.object({
  quantity: z.number().int().min(1).max(99),
});

export const applyCartDiscountInputSchema = z.object({
  code: z.string().min(3).max(64),
});

export const orderStatusSchema = z.enum([
  "pending",
  "confirmed",
  "in_production",
  "ready",
  "cancelled",
  "fulfilled",
]);

export const checkoutInputSchema = z.object({
  pickup_at: apiDateTimeSchema,
});

export const orderItemSchema = z.object({
  id: z.string().uuid(),
  line_type: z.enum(["product", "combo"]),
  product_id: z.string().uuid().nullable().optional(),
  combo_id: z.string().uuid().nullable().optional(),
  name: z.string(),
  slug: catalogSlugSchema,
  quantity: z.number().int().min(1),
  unit_price_cents: z.number().int().min(0),
  line_total_cents: z.number().int().min(0),
});

export const orderSchema = z.object({
  id: z.string().uuid(),
  status: orderStatusSchema,
  subtotal_cents: z.number().int().min(0),
  discount_cents: z.number().int().min(0),
  total_cents: z.number().int().min(0),
  discount_code_snapshot: z.string().nullable().optional(),
  pickup_at: apiDateTimeSchema.nullable().optional(),
  items: z.array(orderItemSchema),
  created_at: apiDateTimeSchema,
  updated_at: apiDateTimeSchema,
});

export const orderSummarySchema = z.object({
  id: z.string().uuid(),
  status: orderStatusSchema,
  total_cents: z.number().int().min(0),
  item_count: z.number().int().min(0),
  pickup_at: apiDateTimeSchema.nullable().optional(),
  created_at: apiDateTimeSchema,
});

export const ordersListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
});

export type Cart = z.infer<typeof cartSchema>;
export type CartItem = z.infer<typeof cartItemSchema>;
export type AddCartItemInput = z.infer<typeof addCartItemInputSchema>;
export type UpdateCartItemInput = z.infer<typeof updateCartItemInputSchema>;
export type ApplyCartDiscountInput = z.infer<typeof applyCartDiscountInputSchema>;
export type CheckoutInput = z.infer<typeof checkoutInputSchema>;
export type OrderStatus = z.infer<typeof orderStatusSchema>;
export type Order = z.infer<typeof orderSchema>;
export type OrderSummary = z.infer<typeof orderSummarySchema>;
export type OrdersListFilterInput = z.infer<typeof ordersListFilterSchema>;

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
