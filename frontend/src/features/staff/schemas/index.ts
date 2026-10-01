import { z } from "zod";

import { catalogSlugSchema } from "@/lib/validation/catalog";
import { USER_ROLE } from "@/constants/roles";
import {
  orderCodeSchema,
  orderPaymentSchema,
  orderStatusSchema,
  orderTimelineEntrySchema,
  orderTypeSchema,
} from "@/lib/schemas/order";
import {
  stationSchema,
  ticketItemSchema,
  ticketStatusSchema,
} from "@/lib/schemas/ticket";
import { apiDateTimeSchema } from "@/lib/validation/datetime";
import { reasonSchema } from "@/lib/validation/reason";

const staffOrderCustomerSchema = z.object({
  user_id: z.uuid(),
  email: z.string().min(1),
  display_name: z.string().nullable().optional(),
  phone: z.string().optional(),
});

const staffOrderItemSchema = z.object({
  id: z.uuid(),
  line_type: z.enum(["product", "combo"]),
  product_id: z.uuid().nullable().optional(),
  combo_id: z.uuid().nullable().optional(),
  name: z.string().min(1),
  slug: catalogSlugSchema,
  quantity: z.number().int().min(1),
  unit_price_cents: z.number().int().min(0),
  line_total_cents: z.number().int().min(0),
});

/** The counter also sees which role made each move; null when the system made it (an unpaid order expiring). */
const staffOrderTimelineEntrySchema = orderTimelineEntrySchema.extend({
  actor_role: z.enum(USER_ROLE).nullable(),
});

/** What each station makes for the order, so the counter can move a ticket nobody started. */
const staffOrderTicketSchema = z.object({
  id: z.uuid(),
  station: stationSchema,
  status: ticketStatusSchema,
  items: z.array(ticketItemSchema),
});

export const staffOrderSummarySchema = z.object({
  id: z.uuid(),
  code: orderCodeSchema,
  status: orderStatusSchema,
  total_cents: z.number().int().min(0),
  item_count: z.number().int().min(0),
  customer: staffOrderCustomerSchema,
  pickup_at: apiDateTimeSchema.nullable().optional(),
  created_at: apiDateTimeSchema,
});

export const staffOrderSchema = z.object({
  id: z.uuid(),
  code: orderCodeSchema,
  status: orderStatusSchema,
  order_type: orderTypeSchema,
  subtotal_cents: z.number().int().min(0),
  discount_cents: z.number().int().min(0),
  total_cents: z.number().int().min(0),
  discount_code_snapshot: z.string().nullable().optional(),
  pickup_at: apiDateTimeSchema.nullable().optional(),
  items: z.array(staffOrderItemSchema),
  timeline: z.array(staffOrderTimelineEntrySchema),
  tickets: z.array(staffOrderTicketSchema),
  customer: staffOrderCustomerSchema,
  payment: orderPaymentSchema.nullable(),
  created_at: apiDateTimeSchema,
  updated_at: apiDateTimeSchema,
});

export const staffOrdersListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
  status: orderStatusSchema.optional(),
});

/** The reason the customer reads when the bakery cancels their order. */
const cancelReasonSchema = reasonSchema("Tell the customer why");

/** A move at the counter; a cancellation states the reason the customer reads. */
export const patchStaffOrderStatusInputSchema = z.union([
  z.object({ status: z.enum(["confirmed", "fulfilled"]) }),
  z.object({ status: z.literal("cancelled"), reason: cancelReasonSchema }),
]);

export const cancelOrderFormSchema = z.object({ reason: cancelReasonSchema });

/** Hand a ticket nobody started to the other station. */
export const moveTicketInputSchema = z.object({
  station: stationSchema,
});

type StaffOrderSummary = z.infer<typeof staffOrderSummarySchema>;
export type StaffOrder = z.infer<typeof staffOrderSchema>;
export type StaffOrdersListFilterInput = z.infer<typeof staffOrdersListFilterSchema>;
export type PatchStaffOrderStatusInput = z.infer<typeof patchStaffOrderStatusInputSchema>;
export type CancelOrderFormInput = z.input<typeof cancelOrderFormSchema>;
export type StaffOrderTicket = z.infer<typeof staffOrderTicketSchema>;
export type MoveTicketInput = z.infer<typeof moveTicketInputSchema>;

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
