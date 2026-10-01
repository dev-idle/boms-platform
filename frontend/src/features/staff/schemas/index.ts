import { z } from "zod";

import { catalogSlugSchema } from "@/lib/validation/catalog";
import { USER_ROLE } from "@/constants/roles";
import { messageSchema } from "@/lib/schemas/message";
import {
  fulfillmentSchema,
  orderChannelSchema,
  orderCodeSchema,
  orderPaymentSchema,
  customizationSchema,
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
import { vietnamPhoneZodString } from "@/lib/validation/phone";
import { reasonSchema } from "@/lib/validation/reason";

/** Who an order is for; a guest's order has no account, so no id or email. */
const staffOrderCustomerSchema = z.object({
  user_id: z.uuid().nullable(),
  email: z.string().min(1).nullable(),
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
  /** How the customer configured it, null for a plain item. */
  customization: customizationSchema.nullable(),
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
  channel: orderChannelSchema,
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
  channel: orderChannelSchema,
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

/** The 4 digits the customer gives at the counter to collect their order. */
const pickupCodeSchema = z.string().regex(/^\d{4}$/, "Enter the 4-digit code");

/**
 * A move at the counter: handing an online order over takes the customer's
 * pickup code, one staff took the cash it is paid with; a cancellation states
 * the reason the customer reads.
 */
export const patchStaffOrderStatusInputSchema = z.union([
  z.object({ status: z.literal("confirmed") }),
  z.object({ status: z.literal("fulfilled"), pickup_code: pickupCodeSchema }),
  z.object({ status: z.literal("fulfilled"), cash_collected: z.literal(true) }),
  z.object({ status: z.literal("cancelled"), reason: cancelReasonSchema }),
]);

export const cancelOrderFormSchema = z.object({ reason: cancelReasonSchema });

export const handoffFormSchema = z.object({ pickup_code: pickupCodeSchema });

/** A page of one bakery day's pickups, soonest first. */
export const staffPickupsFilterSchema = z.object({
  date: z.iso.date(),
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(50),
});

/** The page, with the slot length lateness is measured by: a pickup is late once its slot has ended. */
export const staffPickupsSchema = z.object({
  slot_minutes: z.number().int().min(1),
  pickups: z.array(staffOrderSummarySchema.extend({ pickup_at: apiDateTimeSchema })),
});

/** A product or a combo on an order staff take, with how many. */
export const staffOrderItemInputSchema = z.union([
  z.object({ product_id: z.uuid(), quantity: z.number().int().min(1).max(99) }),
  z.object({ combo_id: z.uuid(), quantity: z.number().int().min(1).max(99) }),
]);

/** POST /staff/orders/quote — what the items cost now and ask of the bakery. */
export const staffOrderQuoteSchema = z.object({
  total_cents: z.number().int().min(0),
  fulfillment: fulfillmentSchema,
});

/** The email of the customer account an order is taken for. */
export const staffCustomerLookupSchema = z.object({
  email: z.email("Enter an email address").trim(),
});

/** Who collects a guest's order: a name and a Vietnam mobile number. */
export const staffGuestFormSchema = z.object({
  name: z.string().trim().min(1, "Enter the guest's name").max(100, "Use at most 100 characters"),
  phone: vietnamPhoneZodString().refine((phone) => Boolean(phone), "Enter the guest's phone number"),
});

/** An order staff take at the counter or on the phone, for an account or a guest. */
export const createStaffOrderInputSchema = z.object({
  channel: z.enum(["counter", "phone"]),
  customer_id: z.uuid().optional(),
  guest: staffGuestFormSchema.optional(),
  pickup_at: apiDateTimeSchema,
  items: z.array(staffOrderItemInputSchema).min(1).max(50),
});

/** GET /staff/customers?email= — the customer account an order is taken for. */
export const staffCustomerSchema = z.object({
  id: z.uuid(),
  email: z.string().min(1),
  display_name: z.string().nullable(),
  phone: z.string().nullable(),
});

/** A product the counter marks sold out for the day. */
export const staffProductSchema = z.object({
  id: z.uuid(),
  name: z.string().min(1),
  category_name: z.string().min(1),
  station: stationSchema,
  sold_out_today: z.boolean(),
});

export const staffProductsFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
  sold_out_today: z.boolean().default(false),
});

/** Whether a conversation still waits on the counter — mirrors backend `conversation.Status`. */
export const conversationStatusSchema = z.enum(["open", "closed"]);

/** One conversation in the counter's inbox, with the start of its last message. */
export const staffInboxConversationSchema = z.object({
  order_id: z.uuid(),
  order_code: orderCodeSchema,
  status: conversationStatusSchema,
  unread: z.number().int().min(0),
  last_message_at: apiDateTimeSchema,
  customer_name: z.string().nullable(),
  customer_email: z.string().min(1),
  preview: z.string(),
});

/** The inbox: a page of open or resolved conversations, or of every one. */
export const staffInboxFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
  status: conversationStatusSchema.optional(),
});

/** GET /staff/conversations/counts — what waits on the counter. */
export const staffConversationCountsSchema = z.object({
  open: z.number().int().min(0),
  unread: z.number().int().min(0),
});

/** Where an order's conversation stands at the counter. */
export const staffConversationSchema = z.object({
  status: conversationStatusSchema,
  unread: z.number().int().min(0),
  assigned_staff_name: z.string().nullable(),
});

/** GET /staff/orders/:id/messages — a page of the order's messages; no conversation before the first. */
export const staffThreadSchema = z.object({
  conversation: staffConversationSchema.nullable(),
  messages: z.array(messageSchema),
  has_more: z.boolean(),
});

/** Hand a ticket nobody started to the other station. */
export const moveTicketInputSchema = z.object({
  station: stationSchema,
});

type StaffOrderSummary = z.infer<typeof staffOrderSummarySchema>;
export type StaffOrderCustomer = z.infer<typeof staffOrderCustomerSchema>;
export type StaffOrder = z.infer<typeof staffOrderSchema>;
export type StaffOrdersListFilterInput = z.infer<typeof staffOrdersListFilterSchema>;
export type PatchStaffOrderStatusInput = z.infer<typeof patchStaffOrderStatusInputSchema>;
export type CancelOrderFormInput = z.input<typeof cancelOrderFormSchema>;
export type HandoffFormInput = z.input<typeof handoffFormSchema>;
export type StaffPickupsFilterInput = z.input<typeof staffPickupsFilterSchema>;
type StaffPickups = z.infer<typeof staffPickupsSchema>;
export type StaffPickup = StaffPickups["pickups"][number];
export type StaffOrderTicket = z.infer<typeof staffOrderTicketSchema>;
export type StaffOrderItemInput = z.infer<typeof staffOrderItemInputSchema>;
export type StaffOrderQuote = z.infer<typeof staffOrderQuoteSchema>;
export type StaffGuestForm = z.infer<typeof staffGuestFormSchema>;
export type StaffCustomerLookup = z.infer<typeof staffCustomerLookupSchema>;
export type CreateStaffOrderInput = z.input<typeof createStaffOrderInputSchema>;
export type StaffCustomer = z.infer<typeof staffCustomerSchema>;
export type StaffProduct = z.infer<typeof staffProductSchema>;
export type StaffProductsFilterInput = z.input<typeof staffProductsFilterSchema>;
export type StaffProductsListResult = {
  products: StaffProduct[];
  pagination: StaffOrdersListResult["pagination"];
};
export type MoveTicketInput = z.infer<typeof moveTicketInputSchema>;
export type ConversationStatus = z.infer<typeof conversationStatusSchema>;
type StaffInboxConversation = z.infer<typeof staffInboxConversationSchema>;
export type StaffInboxFilterInput = z.input<typeof staffInboxFilterSchema>;
export type StaffInboxResult = {
  conversations: StaffInboxConversation[];
  pagination: StaffOrdersListResult["pagination"];
};
export type StaffConversationCounts = z.infer<typeof staffConversationCountsSchema>;
export type StaffThread = z.infer<typeof staffThreadSchema>;
export type StaffConversation = z.infer<typeof staffConversationSchema>;

export type StaffPickupsResult = Omit<StaffPickups, "pickups"> & {
  pickups: StaffPickup[];
  pagination: StaffOrdersListResult["pagination"];
};

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
