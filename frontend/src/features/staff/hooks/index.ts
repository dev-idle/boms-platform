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
  getStaffOrder,
  listStaffOrders,
  patchStaffOrderStatus,
} from "../api";
import {
  staffOrdersListFilterSchema,
  type PatchStaffOrderStatusInput,
  type StaffOrdersListFilterInput,
} from "../schemas";
import { staffQueryKeys } from "./query-options";

export { staffQueryKeys } from "./query-options";

const defaultOrdersFilter: StaffOrdersListFilterInput = {
  page: 1,
  page_size: 20,
};

function staffMutationErrorMessage(error: unknown, fallback: string): string {
  if (isApiError(error)) {
    return error.message;
  }
  return fallback;
}

export function useStaffOrders(
  input: StaffOrdersListFilterInput = defaultOrdersFilter,
) {
  const filter = staffOrdersListFilterSchema.parse(input);
  return useQuery({
    queryKey: staffQueryKeys.orders(filter),
    queryFn: () => listStaffOrders(filter),
    placeholderData: keepPreviousData,
  });
}

export function useStaffOrder(id: string) {
  const isValidId = z.string().uuid().safeParse(id).success;
  return useQuery({
    queryKey: staffQueryKeys.order(id),
    queryFn: () => getStaffOrder(id),
    enabled: isValidId,
    retry: false,
  });
}

export function usePatchStaffOrderStatus(orderId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: PatchStaffOrderStatusInput) =>
      patchStaffOrderStatus(orderId, input),
    onSuccess: (order) => {
      queryClient.setQueryData(staffQueryKeys.order(orderId), order);
      queryClient.invalidateQueries({ queryKey: staffQueryKeys.ordersRoot });
      toast.success("Order status updated");
    },
    onError: (error) => {
      toast.error(
        staffMutationErrorMessage(error, "Failed to update order status"),
      );
    },
  });
}
