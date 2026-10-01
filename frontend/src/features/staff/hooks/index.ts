"use client";

import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useRef } from "react";
import { z } from "zod";
import { toast } from "sonner";

import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import {
  stationTicketsListFilterSchema,
  ticketChangeMessage,
  type PatchTicketStatusInput,
  type StationTicketsListFilterInput,
} from "@/lib/schemas/ticket";

import {
  createStaffOrder,
  findStaffCustomer,
  getStaffOrder,
  listCounterTickets,
  listStaffOrders,
  listStaffPickups,
  listStaffProducts,
  moveTicket,
  patchProductSoldOut,
  quoteStaffOrder,
  patchCounterTicketStatus,
  patchStaffOrderStatus,
} from "../api";
import {
  staffOrdersListFilterSchema,
  staffPickupsFilterSchema,
  staffProductsFilterSchema,
  type CreateStaffOrderInput,
  type MoveTicketInput,
  type StaffOrderItemInput,
  type StaffProductsFilterInput,
  type PatchStaffOrderStatusInput,
  type StaffOrdersListFilterInput,
  type StaffPickupsFilterInput,
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

/** A page of one bakery day's pickups; the last page stays shown while the next loads. */
export function useStaffPickups(input: StaffPickupsFilterInput) {
  const filter = staffPickupsFilterSchema.parse(input);
  return useQuery({
    queryKey: staffQueryKeys.pickups(filter),
    queryFn: () => listStaffPickups(filter),
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

/**
 * Hands a ready order over: an online one with its customer's pickup code, one
 * staff took with the cash it is paid with. The dialog shows a failure, beside
 * the field when the code is wrong or locked.
 */
export function useHandOverOrder(orderId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (proof: { pickup_code: string } | { cash_collected: true }) =>
      patchStaffOrderStatus(orderId, { status: "fulfilled", ...proof }),
    onSuccess: (order) => {
      queryClient.setQueryData(staffQueryKeys.order(orderId), order);
      queryClient.invalidateQueries({ queryKey: staffQueryKeys.ordersRoot });
      toast.success(`${order.code} handed over`);
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

/** What the chosen items cost now and ask of the bakery; the last answer stays while the next loads. */
export function useStaffOrderQuote(items: StaffOrderItemInput[]) {
  return useQuery({
    queryKey: staffQueryKeys.quote(items),
    queryFn: () => quoteStaffOrder(items),
    enabled: items.length > 0,
    placeholderData: keepPreviousData,
    retry: false,
  });
}

/** Takes an order at the counter or on the phone and opens it. */
export function useCreateStaffOrder() {
  const queryClient = useQueryClient();
  const router = useRouter();
  // One key per order being taken: a retry after a lost answer gets back the
  // order the first attempt took.
  const orderKey = useRef(crypto.randomUUID());
  return useMutation({
    mutationFn: (input: CreateStaffOrderInput) => createStaffOrder(input, orderKey.current),
    onSuccess: (order) => {
      orderKey.current = crypto.randomUUID();
      queryClient.invalidateQueries({ queryKey: staffQueryKeys.ordersRoot });
      toast.success(`${order.code} taken`);
      router.push(ROUTE.staff.orderDetail(order.id));
    },
    onError: (error) => {
      toast.error(staffMutationErrorMessage(error, "Failed to take the order"));
    },
  });
}

/** The customer account with that email; empty email looks nothing up. */
export function useStaffCustomer(email: string) {
  return useQuery({
    queryKey: staffQueryKeys.customer(email),
    queryFn: () => findStaffCustomer(email),
    enabled: email !== "",
    retry: false,
  });
}

export function useStaffProducts(input: StaffProductsFilterInput) {
  const filter = staffProductsFilterSchema.parse(input);
  return useQuery({
    queryKey: staffQueryKeys.products(filter),
    queryFn: () => listStaffProducts(filter),
    placeholderData: keepPreviousData,
  });
}

/** Marks a product sold out for today, or back; the list holds several, so the product rides with each call. */
export function useSetProductSoldOut() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ productId, soldOut }: { productId: string; soldOut: boolean }) =>
      patchProductSoldOut(productId, soldOut),
    onSuccess: (product) => {
      queryClient.invalidateQueries({ queryKey: staffQueryKeys.productsRoot });
      toast.success(product.sold_out_today ? `${product.name} is sold out today` : `${product.name} is back`);
    },
    onError: (error) => {
      toast.error(staffMutationErrorMessage(error, "Failed to update the product"));
    },
  });
}
