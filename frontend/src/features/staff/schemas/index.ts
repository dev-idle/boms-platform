import { z } from "zod";

import { catalogSlugSchema } from "@/lib/validation/catalog";
import { apiDateTimeSchema } from "@/lib/validation/datetime";

export const orderStatusSchema = z.enum([
  "pending",
  "confirmed",
  "cancelled",
  "fulfilled",
]);

export const staffOrderCustomerSchema = z.object({
  user_id: z.string().uuid(),
  email: z.string().min(1),
  display_name: z.string().nullable().optional(),
});

export const staffOrderItemSchema = z.object({
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

export const staffOrderSummarySchema = z.object({
  id: z.string().uuid(),
  status: orderStatusSchema,
  total_cents: z.number().int().min(0),
  item_count: z.number().int().min(0),
  customer: staffOrderCustomerSchema,
  created_at: apiDateTimeSchema,
});

export const staffOrderSchema = z.object({
  id: z.string().uuid(),
  status: orderStatusSchema,
  subtotal_cents: z.number().int().min(0),
  discount_cents: z.number().int().min(0),
  total_cents: z.number().int().min(0),
  discount_code_snapshot: z.string().nullable().optional(),
  items: z.array(staffOrderItemSchema),
  customer: staffOrderCustomerSchema,
  created_at: apiDateTimeSchema,
  updated_at: apiDateTimeSchema,
});

export const staffOrdersListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
  status: orderStatusSchema.optional(),
});

export const patchStaffOrderStatusInputSchema = z.object({
  status: z.enum(["confirmed", "cancelled", "fulfilled"]),
});

export type OrderStatus = z.infer<typeof orderStatusSchema>;
export type StaffOrderCustomer = z.infer<typeof staffOrderCustomerSchema>;
export type StaffOrderSummary = z.infer<typeof staffOrderSummarySchema>;
export type StaffOrder = z.infer<typeof staffOrderSchema>;
export type StaffOrdersListFilterInput = z.infer<typeof staffOrdersListFilterSchema>;
export type PatchStaffOrderStatusInput = z.infer<typeof patchStaffOrderStatusInputSchema>;

export type StaffOrdersListResult = {
  orders: StaffOrderSummary[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
};
