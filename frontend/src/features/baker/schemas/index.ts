import { z } from "zod";

import { catalogSlugSchema } from "@/lib/validation/catalog";
import { apiDateTimeSchema } from "@/lib/validation/datetime";

export const bakerOrderStatusSchema = z.enum([
  "confirmed",
  "in_production",
  "ready",
]);

export const bakerOrderCustomerSchema = z.object({
  user_id: z.string().uuid(),
  email: z.string().min(1),
  display_name: z.string().nullable().optional(),
});

export const bakerOrderItemSchema = z.object({
  id: z.string().uuid(),
  line_type: z.enum(["product", "combo"]),
  product_id: z.string().uuid().nullable().optional(),
  combo_id: z.string().uuid().nullable().optional(),
  name: z.string().min(1),
  slug: catalogSlugSchema,
  quantity: z.number().int().min(1),
  unit_price_cents: z.number().int().min(0),
  line_total_cents: z.number().int().min(0),
});

export const bakerOrderSummarySchema = z.object({
  id: z.string().uuid(),
  status: bakerOrderStatusSchema,
  total_cents: z.number().int().min(0),
  item_count: z.number().int().min(0),
  customer: bakerOrderCustomerSchema,
  pickup_at: apiDateTimeSchema.nullable().optional(),
  created_at: apiDateTimeSchema,
});

export const bakerOrderSchema = z.object({
  id: z.string().uuid(),
  status: bakerOrderStatusSchema,
  subtotal_cents: z.number().int().min(0),
  discount_cents: z.number().int().min(0),
  total_cents: z.number().int().min(0),
  discount_code_snapshot: z.string().nullable().optional(),
  pickup_at: apiDateTimeSchema.nullable().optional(),
  items: z.array(bakerOrderItemSchema),
  customer: bakerOrderCustomerSchema,
  created_at: apiDateTimeSchema,
  updated_at: apiDateTimeSchema,
});

export const bakerOrdersListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
  status: bakerOrderStatusSchema.optional(),
});

export const patchBakerOrderStatusInputSchema = z.object({
  status: z.enum(["in_production", "ready"]),
});

export type BakerOrderStatus = z.infer<typeof bakerOrderStatusSchema>;
export type BakerOrderSummary = z.infer<typeof bakerOrderSummarySchema>;
export type BakerOrder = z.infer<typeof bakerOrderSchema>;
export type BakerOrdersListFilterInput = z.infer<typeof bakerOrdersListFilterSchema>;
export type PatchBakerOrderStatusInput = z.infer<typeof patchBakerOrderStatusInputSchema>;

export type BakerOrdersListResult = {
  orders: BakerOrderSummary[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
};
