import { z } from "zod";

import {
  browserRequest,
  browserRequestWithMeta,
} from "@/lib/browser-api-client";
import { parsePaginatedList } from "@/lib/pagination/parse-paginated-list";

import {
  patchStaffOrderStatusInputSchema,
  staffOrderSchema,
  staffOrderSummarySchema,
  staffOrdersListFilterSchema,
  type PatchStaffOrderStatusInput,
  type StaffOrder,
  type StaffOrdersListFilterInput,
  type StaffOrdersListResult,
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

export async function getStaffOrder(id: string): Promise<StaffOrder> {
  const parsedId = z.string().uuid().parse(id);
  return browserRequest<StaffOrder>(`/api/v1/staff/orders/${parsedId}`, {
    method: "GET",
    schema: staffOrderSchema,
  });
}

export async function patchStaffOrderStatus(
  id: string,
  input: PatchStaffOrderStatusInput,
): Promise<StaffOrder> {
  const parsedId = z.string().uuid().parse(id);
  const body = patchStaffOrderStatusInputSchema.parse(input);
  return browserRequest<StaffOrder>(`/api/v1/staff/orders/${parsedId}/status`, {
    method: "PATCH",
    schema: staffOrderSchema,
    json: body,
  });
}
