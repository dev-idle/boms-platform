"use client";

import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, useRef } from "react";
import { z } from "zod";
import { toast } from "sonner";

import { ROUTE } from "@/constants/routes";
import { ApiErrorCode, isApiError } from "@/lib/errors";

import {
  addCartItem,
  applyCartDiscount,
  cancelOrder,
  capturePayment,
  checkoutCart,
  getCart,
  getOrder,
  getPickupRules,
  getPickupSlots,
  listOrders,
  removeCartDiscount,
  removeCartItem,
  rescheduleOrder,
  startPayment,
  updateCartItem,
} from "../api";
import {
  ordersListFilterSchema,
  type AddCartItemInput,
  type ApplyCartDiscountInput,
  type CheckoutInput,
  type OrdersListFilterInput,
  type RescheduleInput,
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
  const router = useRouter();
  // One key per order being placed: a retry after a lost answer gets back the
  // order the first attempt placed, instead of an empty cart.
  const checkoutKey = useRef(crypto.randomUUID());
  return useMutation({
    mutationFn: (input: CheckoutInput) => checkoutCart(input, checkoutKey.current),
    // Here, not in the caller: the order's realtime notice can empty the cart
    // and unmount the checkout panel before this answer arrives.
    onSuccess: (order) => {
      checkoutKey.current = crypto.randomUUID();
      queryClient.invalidateQueries({ queryKey: customerQueryKeys.cart });
      queryClient.invalidateQueries({ queryKey: customerQueryKeys.ordersRoot });
      toast.success("Order placed. Pay with PayPal to confirm it.");
      router.push(ROUTE.orderDetail(order.id));
    },
    onError: (error) => {
      // The rules, the slots, or what the cart's items need (a manager may have
      // moved a category or changed a lead time) may be older than checkout's.
      if (isApiError(error) && PICKUP_REFUSALS.has(error.code)) {
        void queryClient.invalidateQueries({ queryKey: customerQueryKeys.pickupRules });
        void queryClient.invalidateQueries({ queryKey: customerQueryKeys.pickupSlotsRoot });
        void queryClient.invalidateQueries({ queryKey: customerQueryKeys.cart });
      }
      toast.error(
        isApiError(error) && error.code === ApiErrorCode.TermsNotAccepted
          ? "Our policies were just updated. Reload the page, then read and accept the current version."
          : isApiError(error) && error.code === ApiErrorCode.EmailNotVerified
            ? "Confirm your email address before placing an order. We sent you a link."
            : cartMutationErrorMessage(error, "Checkout failed"),
      );
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

/** Sends the customer to PayPal to approve the order's payment. */
export function useStartPayment(orderId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => startPayment(orderId),
    onSuccess: (approveURL) => {
      window.location.assign(approveURL);
    },
    onError: (error) => {
      if (isApiError(error) && error.code === ApiErrorCode.OrderNotPayable) {
        void queryClient.invalidateQueries({ queryKey: customerQueryKeys.order(orderId) });
        toast.error("This order can no longer be paid.");
        return;
      }
      toast.error("We could not open PayPal. Please try again.");
    },
  });
}

function useCapturePayment(orderId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => capturePayment(orderId),
    onSuccess: ({ status }) => {
      void queryClient.invalidateQueries({ queryKey: customerQueryKeys.order(orderId) });
      void queryClient.invalidateQueries({ queryKey: customerQueryKeys.ordersRoot });
      toast.success(
        status === "captured"
          ? "Payment received. Your order is confirmed."
          : "PayPal is reviewing your payment. Your order is confirmed as soon as it clears.",
      );
    },
    onError: (error) => {
      if (isApiError(error) && error.code === ApiErrorCode.OrderNotPayable) {
        // Back from PayPal after the time to pay ran out: nothing was taken.
        toast.error("This order can no longer be paid. Nothing was charged.");
        return;
      }
      toast.error(
        isApiError(error) && error.code === ApiErrorCode.PaymentNotCompleted
          ? "The payment was not completed. Try again, or choose another way to pay on PayPal."
          : "We could not confirm your payment. Reload the page to see where it stands.",
      );
    },
  });
}

/**
 * Captures the payment once when the customer comes back from approving it on
 * PayPal (`?paypal=approved`), and takes PayPal's parameters out of the
 * address bar so a reload does not capture again.
 */
export function usePayPalReturn(orderId: string) {
  const capture = useCapturePayment(orderId);
  const { mutate } = capture;
  // A second effect run (Strict Mode) finds the parameters already gone.
  const handled = useRef(false);

  useEffect(() => {
    if (handled.current) {
      return;
    }
    handled.current = true;
    if (new URLSearchParams(window.location.search).get("paypal") !== "approved") {
      return;
    }
    window.history.replaceState(null, "", window.location.pathname);
    mutate();
  }, [mutate]);

  return capture;
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

/** Cancels the customer's order while the bakery has not started it. */
export function useCancelOrder(orderId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => cancelOrder(orderId),
    onSuccess: (order) => {
      void queryClient.invalidateQueries({ queryKey: customerQueryKeys.order(orderId) });
      void queryClient.invalidateQueries({ queryKey: customerQueryKeys.ordersRoot });
      toast.success(
        order.payment?.refund_requested_at
          ? "Order cancelled. Your refund is on its way to PayPal."
          : "Order cancelled.",
      );
    },
    onError: (error) => {
      if (isApiError(error) && error.code === ApiErrorCode.InvalidOrderStatusTransition) {
        void queryClient.invalidateQueries({ queryKey: customerQueryKeys.order(orderId) });
        toast.error("We have started making this order, so it can no longer be cancelled.");
        return;
      }
      if (isApiError(error) && error.code === ApiErrorCode.PaymentUnderReview) {
        void queryClient.invalidateQueries({ queryKey: customerQueryKeys.order(orderId) });
        toast.error("PayPal is still reviewing your payment. You can cancel once it clears.");
        return;
      }
      toast.error("We could not cancel the order. Please try again.");
    },
  });
}

/** Moves the pickup of the customer's order while the bakery has not started it. */
export function useRescheduleOrder(orderId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: RescheduleInput) => rescheduleOrder(orderId, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: customerQueryKeys.order(orderId) });
      void queryClient.invalidateQueries({ queryKey: customerQueryKeys.ordersRoot });
      toast.success("Pickup time changed.");
    },
    onError: (error) => {
      if (isApiError(error) && error.code === ApiErrorCode.InvalidOrderStatusTransition) {
        void queryClient.invalidateQueries({ queryKey: customerQueryKeys.order(orderId) });
        toast.error("We have started making this order, so its pickup can no longer be changed.");
        return;
      }
      // The rules or the slots may be older than the server's.
      if (isApiError(error) && PICKUP_REFUSALS.has(error.code)) {
        void queryClient.invalidateQueries({ queryKey: customerQueryKeys.pickupRules });
        void queryClient.invalidateQueries({ queryKey: customerQueryKeys.pickupSlotsRoot });
      }
      toast.error(cartMutationErrorMessage(error, "We could not change the pickup time."));
    },
  });
}
