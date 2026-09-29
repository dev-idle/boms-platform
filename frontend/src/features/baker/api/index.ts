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

import { kitchenTicketSchema, type KitchenTicket } from "../schemas";

export async function listKitchenTickets(
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
    `/api/v1/baker/tickets?${params.toString()}`,
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

export async function getKitchenTicket(id: string): Promise<KitchenTicket> {
  const parsedId = z.uuid().parse(id);
  return browserRequest<KitchenTicket>(`/api/v1/baker/tickets/${parsedId}`, {
    method: "GET",
    schema: kitchenTicketSchema,
  });
}

export async function patchKitchenTicketStatus(
  id: string,
  input: PatchTicketStatusInput,
): Promise<TicketChange> {
  const parsedId = z.uuid().parse(id);
  const body = patchTicketStatusInputSchema.parse(input);
  return browserRequest<TicketChange>(`/api/v1/baker/tickets/${parsedId}/status`, {
    method: "PATCH",
    schema: ticketChangeSchema,
    json: body,
  });
}
