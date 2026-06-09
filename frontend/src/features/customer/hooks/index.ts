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
  addCartItem,
  applyCartDiscount,
  checkoutCart,
  getCart,
  getOrder,
  listOrders,
  removeCartDiscount,
  removeCartItem,
  updateCartItem,
} from "../api";
import {
  ordersListFilterSchema,
  type AddCartItemInput,
  type ApplyCartDiscountInput,
  type OrdersListFilterInput,
  type UpdateCartItemInput,
} from "../schemas";
import { customerQueryKeys } from "./query-options";

export { customerQueryKeys } from "./query-options";

function cartMutationErrorMessage(error: unknown, fallback: string): string {
  if (isApiError(error)) {
    return error.message;
  }
  return fallback;
}

export function useCart() {
  return useQuery({
    queryKey: customerQueryKeys.cart,
    queryFn: getCart,
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

export function useCheckoutCart() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: checkoutCart,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: customerQueryKeys.cart });
      queryClient.invalidateQueries({ queryKey: customerQueryKeys.ordersRoot });
      toast.success("Order placed");
    },
    onError: (error) => {
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
  const isValidId = z.string().uuid().safeParse(id).success;
  return useQuery({
    queryKey: customerQueryKeys.order(id),
    queryFn: () => getOrder(id),
    enabled: isValidId,
    retry: false,
  });
}
