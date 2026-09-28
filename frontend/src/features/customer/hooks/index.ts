"use client";

import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { z } from "zod";
import { toast } from "sonner";

import { ApiErrorCode, isApiError } from "@/lib/errors";

import {
  addCartItem,
  applyCartDiscount,
  checkoutCart,
  getCart,
  getOrder,
  getPickupRules,
  getPickupSlots,
  listOrders,
  removeCartDiscount,
  removeCartItem,
  updateCartItem,
} from "../api";
import {
  ordersListFilterSchema,
  type AddCartItemInput,
  type ApplyCartDiscountInput,
  type CheckoutInput,
  type OrdersListFilterInput,
  type UpdateCartItemInput,
} from "../schemas";
import { customerQueryKeys } from "./query-options";

function cartMutationErrorMessage(error: unknown, fallback: string): string {
  if (isApiError(error)) {
    return error.message;
  }
  return fallback;
}

type UseCartOptions = {
  /** Skip fetch when false — use on public chrome for guests (avoids 401 → login redirect). */
  enabled?: boolean;
};

export function useCart(options: UseCartOptions = {}) {
  const { enabled = true } = options;
  return useQuery({
    queryKey: customerQueryKeys.cart,
    queryFn: getCart,
    enabled,
  });
}

export function useAddCartItem() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: AddCartItemInput) => addCartItem(input),
    onSuccess: (cart) => {
      queryClient.setQueryData(customerQueryKeys.cart, cart);
      toast.success("Added to cart");
    },
    onError: (error) => {
      toast.error(cartMutationErrorMessage(error, "Failed to add to cart"));
    },
  });
}

export function useUpdateCartItem(itemId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateCartItemInput) => updateCartItem(itemId, input),
    onSuccess: (cart) => {
      queryClient.setQueryData(customerQueryKeys.cart, cart);
    },
    onError: (error) => {
      toast.error(cartMutationErrorMessage(error, "Failed to update cart item"));
    },
  });
}

export function useRemoveCartItem() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (itemId: string) => removeCartItem(itemId),
    onSuccess: (cart) => {
      queryClient.setQueryData(customerQueryKeys.cart, cart);
      toast.success("Item removed");
    },
    onError: (error) => {
      toast.error(cartMutationErrorMessage(error, "Failed to remove cart item"));
    },
  });
}

export function useApplyCartDiscount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: ApplyCartDiscountInput) => applyCartDiscount(input),
    onSuccess: (cart) => {
      queryClient.setQueryData(customerQueryKeys.cart, cart);
      toast.success("Discount applied");
    },
    onError: (error) => {
      toast.error(cartMutationErrorMessage(error, "Failed to apply discount"));
    },
  });
}

export function useRemoveCartDiscount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: removeCartDiscount,
    onSuccess: (cart) => {
      queryClient.setQueryData(customerQueryKeys.cart, cart);
      toast.success("Discount removed");
    },
    onError: (error) => {
      toast.error(cartMutationErrorMessage(error, "Failed to remove discount"));
    },
  });
}

const PICKUP_REFUSALS: ReadonlySet<string> = new Set([
  ApiErrorCode.PickupTooSoon,
  ApiErrorCode.PickupTooFar,
  ApiErrorCode.PickupClosedDay,
  ApiErrorCode.PickupOutsideHours,
  ApiErrorCode.PickupOffSlot,
  ApiErrorCode.PickupSlotFull,
  ApiErrorCode.PickupDayLimit,
]);

export function useCheckoutCart() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CheckoutInput) => checkoutCart(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: customerQueryKeys.cart });
      queryClient.invalidateQueries({ queryKey: customerQueryKeys.ordersRoot });
      toast.success("Order placed");
    },
    onError: (error) => {
      // The rules, the slots, or what the cart's items need (a manager may have
      // moved a category or changed a lead time) may be older than checkout's.
      if (isApiError(error) && PICKUP_REFUSALS.has(error.code)) {
        void queryClient.invalidateQueries({ queryKey: customerQueryKeys.pickupRules });
        void queryClient.invalidateQueries({ queryKey: customerQueryKeys.pickupSlotsRoot });
        void queryClient.invalidateQueries({ queryKey: customerQueryKeys.cart });
      }
      toast.error(cartMutationErrorMessage(error, "Checkout failed"));
    },
  });
}

const defaultOrdersFilter: OrdersListFilterInput = {
  page: 1,
  page_size: 20,
};

export function useOrders(input: OrdersListFilterInput = defaultOrdersFilter) {
  const filter = ordersListFilterSchema.parse(input);
  return useQuery({
    queryKey: customerQueryKeys.orders(filter),
    queryFn: () => listOrders(filter),
    placeholderData: keepPreviousData,
  });
}

export function useOrder(id: string) {
  const isValidId = z.uuid().safeParse(id).success;
  return useQuery({
    queryKey: customerQueryKeys.order(id),
    queryFn: () => getOrder(id),
    enabled: isValidId,
    retry: false,
  });
}

/** A bakery day's pickup slots; refreshed live as orders take and free them. */
export function usePickupSlots(date: string) {
  return useQuery({
    queryKey: customerQueryKeys.pickupSlots(date),
    queryFn: () => getPickupSlots(date),
    enabled: date !== "",
    placeholderData: keepPreviousData,
  });
}

/** The pickup window checkout enforces; refreshed live when an admin edits it. */
export function usePickupRules() {
  return useQuery({
    queryKey: customerQueryKeys.pickupRules,
    queryFn: getPickupRules,
  });
}
