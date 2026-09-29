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
  stationTicketsListFilterSchema,
  ticketChangeMessage,
  type PatchTicketStatusInput,
  type StationTicketsListFilterInput,
} from "@/lib/schemas/ticket";

import {
  getStaffOrder,
  listCounterTickets,
  listStaffOrders,
  moveTicket,
  patchCounterTicketStatus,
  patchStaffOrderStatus,
} from "../api";
import {
  staffOrdersListFilterSchema,
  type MoveTicketInput,
  type PatchStaffOrderStatusInput,
  type StaffOrdersListFilterInput,
} from "../schemas";
import { staffQueryKeys } from "./query-options";

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
  const isValidId = z.uuid().safeParse(id).success;
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

const defaultTicketsFilter: StationTicketsListFilterInput = {
  page: 1,
  page_size: 20,
};

export function useCounterTickets(
  input: StationTicketsListFilterInput = defaultTicketsFilter,
) {
  const filter = stationTicketsListFilterSchema.parse(input);
  return useQuery({
    queryKey: staffQueryKeys.tickets(filter),
    queryFn: () => listCounterTickets(filter),
    placeholderData: keepPreviousData,
  });
}

/** Starts or finishes one counter ticket; the queue holds several, so the ticket rides with each call. */
export function usePatchCounterTicketStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ ticketId, ...input }: PatchTicketStatusInput & { ticketId: string }) =>
      patchCounterTicketStatus(ticketId, input),
    onSuccess: (change) => {
      queryClient.invalidateQueries({ queryKey: staffQueryKeys.ticketsRoot });
      queryClient.invalidateQueries({ queryKey: staffQueryKeys.order(change.order_id) });
      toast.success(ticketChangeMessage(change));
    },
    onError: (error) => {
      toast.error(staffMutationErrorMessage(error, "Failed to update the ticket"));
    },
  });
}

export function useMoveTicket(orderId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ ticketId, ...input }: MoveTicketInput & { ticketId: string }) =>
      moveTicket(ticketId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: staffQueryKeys.order(orderId) });
      queryClient.invalidateQueries({ queryKey: staffQueryKeys.ticketsRoot });
      toast.success("Ticket moved");
    },
    onError: (error) => {
      toast.error(staffMutationErrorMessage(error, "Failed to move the ticket"));
    },
  });
}
