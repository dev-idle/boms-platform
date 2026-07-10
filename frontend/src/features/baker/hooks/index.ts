"use client";

import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { z } from "zod";
import { toast } from "sonner";

import { isApiError } from "@/lib/errors";

import {
  getBakerProductionOrder,
  listBakerProductionOrders,
  patchBakerOrderStatus,
} from "../api";
import {
  bakerOrdersListFilterSchema,
  type BakerOrdersListFilterInput,
  type PatchBakerOrderStatusInput,
} from "../schemas";
import { bakerQueryKeys } from "./query-options";

export { bakerQueryKeys } from "./query-options";

const defaultOrdersFilter: BakerOrdersListFilterInput = {
  page: 1,
  page_size: 20,
};

function bakerMutationErrorMessage(error: unknown, fallback: string): string {
  if (isApiError(error)) {
    return error.message;
  }
  return fallback;
}

export function useBakerProductionOrders(
  input: BakerOrdersListFilterInput = defaultOrdersFilter,
) {
  const filter = bakerOrdersListFilterSchema.parse(input);
  return useQuery({
    queryKey: bakerQueryKeys.production(filter),
    queryFn: () => listBakerProductionOrders(filter),
    placeholderData: keepPreviousData,
  });
}

export function useBakerProductionOrder(id: string) {
  const isValidId = z.string().uuid().safeParse(id).success;
  return useQuery({
    queryKey: bakerQueryKeys.productionOrder(id),
    queryFn: () => getBakerProductionOrder(id),
    enabled: isValidId,
    retry: false,
  });
}

export function usePatchBakerOrderStatus(orderId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: PatchBakerOrderStatusInput) =>
      patchBakerOrderStatus(orderId, input),
    onSuccess: (order) => {
      queryClient.setQueryData(bakerQueryKeys.productionOrder(orderId), order);
      queryClient.invalidateQueries({ queryKey: bakerQueryKeys.productionRoot });
      toast.success("Production status updated");
    },
    onError: (error) => {
      toast.error(
        bakerMutationErrorMessage(error, "Failed to update production status"),
      );
    },
  });
}
