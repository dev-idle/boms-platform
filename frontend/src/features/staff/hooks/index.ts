"use client";

import {
  keepPreviousData,
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
  type QueryClient,
} from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useRef } from "react";
import { z } from "zod";
import { toast } from "sonner";

import { ROUTE } from "@/constants/routes";
import { pickupQueryKeys } from "@/features/customer";
import { isApiError, userFacingErrorMessage } from "@/lib/errors";
import { MESSAGE_REFUSED_TEXT, type MessageInput } from "@/lib/schemas/message";
import {
  ticketChangeMessage,
  type PatchTicketStatusInput,
} from "@/lib/schemas/ticket";

import {
  createStaffOrder,
  findStaffCustomer,
  getStaffConversationCounts,
  getStaffOrder,
  getStaffMessages,
  listStaffOrders,
  listStaffPickups,
  listStaffConversations,
  listStaffProducts,
  markStaffMessagesRead,
  moveTicket,
  patchProductSoldOut,
  quoteStaffOrder,
  patchCounterTicketStatus,
  patchStaffConversation,
  moveStaffOrderPickup,
  patchStaffOrderStatus,
  postStaffMessage,
  reportOrderIncident,
} from "../api";
import {
  staffInboxFilterSchema,
  staffOrdersListFilterSchema,
  staffPickupsFilterSchema,
  staffProductsFilterSchema,
  type ConversationStatus,
  type CreateStaffOrderInput,
  type MoveStaffOrderPickupInput,
  type MoveTicketInput,
  type StaffOrderItemInput,
  type StaffProductsFilterInput,
  type PatchStaffOrderStatusInput,
  type ReportIncidentInput,
  type StaffOrdersListFilterInput,
  type StaffPickupsFilterInput,
  type StaffInboxFilterInput,
  type StaffThreadRef,
} from "../schemas";
import { staffQueryKeys } from "./query-options";

/** A slot taken, or refused as full, changes what the picker may offer next. */
function invalidatePickupSlots(queryClient: QueryClient): void {
  for (const queryKey of pickupQueryKeys) {
    void queryClient.invalidateQueries({ queryKey });
  }
}

export function useStaffOrders(input: StaffOrdersListFilterInput) {
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

/** Reports what went wrong with an order; the dialog shows a failure itself. */
export function useReportOrderIncident(orderId: string) {
  return useMutation({
    mutationFn: (input: ReportIncidentInput) => reportOrderIncident(orderId, input),
    onSuccess: () => {
      toast.success("Issue reported to the manager");
    },
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
        userFacingErrorMessage(error, "We could not update the order. Please try again."),
      );
    },
  });
}

/** Moves an order's pickup to another free slot ahead. */
export function useMoveStaffOrderPickup(orderId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: MoveStaffOrderPickupInput) => moveStaffOrderPickup(orderId, input),
    onSuccess: (order) => {
      queryClient.setQueryData(staffQueryKeys.order(orderId), order);
      queryClient.invalidateQueries({ queryKey: staffQueryKeys.ordersRoot });
      invalidatePickupSlots(queryClient);
      toast.success("Pickup moved");
    },
    onError: (error) => {
      invalidatePickupSlots(queryClient);
      toast.error(userFacingErrorMessage(error, "We could not move the pickup. Please try again."));
    },
  });
}

/**
 * Hands a ready or missed order over: an online one with its customer's pickup
 * code, one staff took with the cash it is paid with. The dialog shows a failure, beside
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

/** Gets one counter ticket ready; the schedule holds several, so the ticket rides with each call. */
export function usePatchCounterTicketStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ ticketId, ...input }: PatchTicketStatusInput & { ticketId: string }) =>
      patchCounterTicketStatus(ticketId, input),
    onSuccess: (change) => {
      queryClient.invalidateQueries({ queryKey: staffQueryKeys.pickupsRoot });
      queryClient.invalidateQueries({ queryKey: staffQueryKeys.order(change.order_id) });
      toast.success(ticketChangeMessage(change));
    },
    onError: (error) => {
      toast.error(userFacingErrorMessage(error, "We could not update the ticket. Please try again."));
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
      queryClient.invalidateQueries({ queryKey: staffQueryKeys.pickupsRoot });
      toast.success("Ticket moved");
    },
    onError: (error) => {
      toast.error(userFacingErrorMessage(error, "We could not move the ticket. Please try again."));
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
      invalidatePickupSlots(queryClient);
      toast.success(`${order.code} taken`);
      router.push(ROUTE.staff.orderDetail(order.id));
    },
    onError: (error) => {
      invalidatePickupSlots(queryClient);
      toast.error(userFacingErrorMessage(error, "We could not take the order. Please try again."));
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
      toast.error(userFacingErrorMessage(error, "We could not update the product. Please try again."));
    },
  });
}

/** A page of the counter's inbox; refreshed live as messages arrive. */
export function useStaffConversations(input: StaffInboxFilterInput) {
  const filter = staffInboxFilterSchema.parse(input);
  return useQuery({
    queryKey: staffQueryKeys.conversations(filter),
    queryFn: () => listStaffConversations(filter),
    placeholderData: keepPreviousData,
  });
}

/** What waits on the counter: open conversations and those holding unread messages. */
export function useStaffConversationCounts() {
  return useQuery({
    queryKey: staffQueryKeys.conversationCounts,
    queryFn: getStaffConversationCounts,
  });
}

/** A conversation's messages, the latest page first; each older page starts before the oldest loaded. */
export function useStaffMessages(thread: StaffThreadRef) {
  return useInfiniteQuery({
    queryKey: staffQueryKeys.messages(thread),
    queryFn: ({ pageParam }) => getStaffMessages(thread, pageParam),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (page) => (page.has_more ? page.messages[0]?.id : undefined),
    retry: false,
  });
}

/** Writes a staff member's message to the conversation's customer. */
export function usePostStaffMessage(thread: StaffThreadRef) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: MessageInput) => postStaffMessage(thread, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: staffQueryKeys.messages(thread) });
      void queryClient.invalidateQueries({ queryKey: staffQueryKeys.conversationsRoot });
    },
    onError: (error) => {
      toast.error(
        isApiError(error) && error.hasValidationDetails()
          ? MESSAGE_REFUSED_TEXT
          : userFacingErrorMessage(error, "We could not send the message. Please try again."),
      );
    },
  });
}

/** Records that the counter read a conversation's messages. */
export function useMarkStaffMessagesRead(thread: StaffThreadRef) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => markStaffMessagesRead(thread),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: staffQueryKeys.messages(thread) });
      void queryClient.invalidateQueries({ queryKey: staffQueryKeys.conversationsRoot });
    },
  });
}

/** Resolves a conversation, or opens it again. */
export function useSetConversationStatus(thread: StaffThreadRef) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (status: ConversationStatus) => patchStaffConversation(thread, status),
    onSuccess: (conversation) => {
      void queryClient.invalidateQueries({ queryKey: staffQueryKeys.messages(thread) });
      void queryClient.invalidateQueries({ queryKey: staffQueryKeys.conversationsRoot });
      toast.success(conversation.status === "closed" ? "Conversation resolved" : "Conversation opened again");
    },
    onError: (error) => {
      toast.error(userFacingErrorMessage(error, "We could not update the conversation. Please try again."));
    },
  });
}
