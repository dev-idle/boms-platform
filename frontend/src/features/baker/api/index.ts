import { z } from "zod";

import {
  browserRequest,
  browserRequestWithMeta,
} from "@/lib/browser-api-client";
import { parsePaginatedList } from "@/lib/pagination/parse-paginated-list";

import {
  bakerOrderSchema,
  bakerOrderSummarySchema,
  bakerOrdersListFilterSchema,
  patchBakerOrderStatusInputSchema,
  type BakerOrder,
  type BakerOrdersListFilterInput,
  type BakerOrdersListResult,
  type PatchBakerOrderStatusInput,
} from "../schemas";

export async function listBakerProductionOrders(
  input: BakerOrdersListFilterInput,
): Promise<BakerOrdersListResult> {
  const filter = bakerOrdersListFilterSchema.parse(input);
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  if (filter.status) {
    params.set("status", filter.status);
  }
  const result = await browserRequestWithMeta<z.infer<typeof bakerOrderSummarySchema>[]>(
    `/api/v1/baker/production?${params.toString()}`,
    { method: "GET", schema: z.array(bakerOrderSummarySchema) },
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

export async function getBakerProductionOrder(id: string): Promise<BakerOrder> {
  const parsedId = z.string().uuid().parse(id);
  return browserRequest<BakerOrder>(`/api/v1/baker/production/${parsedId}`, {
    method: "GET",
    schema: bakerOrderSchema,
  });
}

export async function patchBakerOrderStatus(
  id: string,
  input: PatchBakerOrderStatusInput,
): Promise<BakerOrder> {
  const parsedId = z.string().uuid().parse(id);
  const body = patchBakerOrderStatusInputSchema.parse(input);
  return browserRequest<BakerOrder>(`/api/v1/baker/production/${parsedId}/status`, {
    method: "PATCH",
    schema: bakerOrderSchema,
    json: body,
  });
}
