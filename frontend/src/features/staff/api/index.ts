import { z } from "zod";

import {
  browserRequest,
  browserRequestWithMeta,
} from "@/lib/browser-api-client";
import { parsePaginatedList } from "@/lib/pagination/parse-paginated-list";
import {
  patchTicketStatusInputSchema,
  stationTicketSchema,
  stationTicketsListFilterSchema,
  ticketChangeSchema,
  type PatchTicketStatusInput,
  type StationTicket,
  type StationTicketsListFilterInput,
  type StationTicketsListResult,
  type TicketChange,
} from "@/lib/schemas/ticket";

import {
  moveTicketInputSchema,
  patchStaffOrderStatusInputSchema,
  staffOrderSchema,
  staffOrderSummarySchema,
  staffOrdersListFilterSchema,
  staffPickupsFilterSchema,
  staffPickupsSchema,
  type MoveTicketInput,
  type PatchStaffOrderStatusInput,
  type StaffOrder,
  type StaffOrdersListFilterInput,
  type StaffOrdersListResult,
  type StaffPickupsFilterInput,
  type StaffPickupsResult,
} from "../schemas";

export async function listStaffOrders(
  input: StaffOrdersListFilterInput,
): Promise<StaffOrdersListResult> {
  const filter = staffOrdersListFilterSchema.parse(input);
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  if (filter.status) {
    params.set("status", filter.status);
  }
  const result = await browserRequestWithMeta<z.infer<typeof staffOrderSummarySchema>[]>(
    `/api/v1/staff/orders?${params.toString()}`,
    { method: "GET", schema: z.array(staffOrderSummarySchema) },
  );
  const parsed = parsePaginatedList(result.data, result.meta, {
    page: filter.page,
    page_size: filter.page_size,
  });
  return {
    orders: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}

/** A page of one bakery day's pickups. */
export async function listStaffPickups(input: StaffPickupsFilterInput): Promise<StaffPickupsResult> {
  const filter = staffPickupsFilterSchema.parse(input);
  const params = new URLSearchParams({
    date: filter.date,
    page: String(filter.page),
    page_size: String(filter.page_size),
  });
  const result = await browserRequestWithMeta<z.infer<typeof staffPickupsSchema>>(
    `/api/v1/staff/pickups?${params.toString()}`,
    { method: "GET", schema: staffPickupsSchema },
  );
  const parsed = parsePaginatedList(result.data.pickups, result.meta, {
    page: filter.page,
    page_size: filter.page_size,
  });
  return {
    slot_minutes: result.data.slot_minutes,
    pickups: parsed.items,
    pagination: parsed.pagination,
  };
}

export async function getStaffOrder(id: string): Promise<StaffOrder> {
  const parsedId = z.uuid().parse(id);
  return browserRequest<StaffOrder>(`/api/v1/staff/orders/${parsedId}`, {
    method: "GET",
    schema: staffOrderSchema,
  });
}

export async function patchStaffOrderStatus(
  id: string,
  input: PatchStaffOrderStatusInput,
): Promise<StaffOrder> {
  const parsedId = z.uuid().parse(id);
  const body = patchStaffOrderStatusInputSchema.parse(input);
  return browserRequest<StaffOrder>(`/api/v1/staff/orders/${parsedId}/status`, {
    method: "PATCH",
    schema: staffOrderSchema,
    json: body,
  });
}

export async function listCounterTickets(
  input: StationTicketsListFilterInput,
): Promise<StationTicketsListResult> {
  const filter = stationTicketsListFilterSchema.parse(input);
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  if (filter.status) {
    params.set("status", filter.status);
  }
  const result = await browserRequestWithMeta<StationTicket[]>(
    `/api/v1/staff/tickets?${params.toString()}`,
    { method: "GET", schema: z.array(stationTicketSchema) },
  );
  const parsed = parsePaginatedList(result.data, result.meta, {
    page: filter.page,
    page_size: filter.page_size,
  });
  return {
    tickets: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}

export async function patchCounterTicketStatus(
  id: string,
  input: PatchTicketStatusInput,
): Promise<TicketChange> {
  const parsedId = z.uuid().parse(id);
  const body = patchTicketStatusInputSchema.parse(input);
  return browserRequest<TicketChange>(`/api/v1/staff/tickets/${parsedId}/status`, {
    method: "PATCH",
    schema: ticketChangeSchema,
    json: body,
  });
}

export async function moveTicket(id: string, input: MoveTicketInput): Promise<TicketChange> {
  const parsedId = z.uuid().parse(id);
  const body = moveTicketInputSchema.parse(input);
  return browserRequest<TicketChange>(`/api/v1/staff/tickets/${parsedId}/station`, {
    method: "PATCH",
    schema: ticketChangeSchema,
    json: body,
  });
}
