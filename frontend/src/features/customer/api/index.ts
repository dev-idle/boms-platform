import { z } from "zod";

import { TERMS_VERSION } from "@/constants/policies";
import {
  browserRequest,
  browserRequestWithMeta,
} from "@/lib/browser-api-client";
import { parsePaginatedList } from "@/lib/pagination/parse-paginated-list";

import {
  addCartItemInputSchema,
  applyCartDiscountInputSchema,
  cartSchema,
  checkoutInputSchema,
  orderSchema,
  orderSummarySchema,
  ordersListFilterSchema,
  paymentCaptureSchema,
  paymentStartSchema,
  pickupRulesSchema,
  pickupSlotsSchema,
  updateCartItemInputSchema,
  type AddCartItemInput,
  type ApplyCartDiscountInput,
  type Cart,
  type CheckoutInput,
  type Order,
  type OrdersListFilterInput,
  type OrdersListResult,
  type PaymentCapture,
  type PaymentStart,
  type PickupRules,
  type PickupSlots,
  type UpdateCartItemInput,
} from "../schemas";

export async function getCart(): Promise<Cart> {
  return browserRequest<Cart>("/api/v1/cart", {
    method: "GET",
    schema: cartSchema,
  });
}

export async function addCartItem(input: AddCartItemInput): Promise<Cart> {
  const body = addCartItemInputSchema.parse(input);
  return browserRequest<Cart>("/api/v1/cart/items", {
    method: "POST",
    schema: cartSchema,
    json: body,
  });
}

export async function updateCartItem(
  itemId: string,
  input: UpdateCartItemInput,
): Promise<Cart> {
  const id = z.uuid().parse(itemId);
  const body = updateCartItemInputSchema.parse(input);
  return browserRequest<Cart>(`/api/v1/cart/items/${id}`, {
    method: "PATCH",
    schema: cartSchema,
    json: body,
  });
}

export async function removeCartItem(itemId: string): Promise<Cart> {
  const id = z.uuid().parse(itemId);
  return browserRequest<Cart>(`/api/v1/cart/items/${id}`, {
    method: "DELETE",
    schema: cartSchema,
  });
}

export async function applyCartDiscount(
  input: ApplyCartDiscountInput,
): Promise<Cart> {
  const body = applyCartDiscountInputSchema.parse(input);
  return browserRequest<Cart>("/api/v1/cart/discount", {
    method: "PUT",
    schema: cartSchema,
    json: body,
  });
}

export async function removeCartDiscount(): Promise<Cart> {
  return browserRequest<Cart>("/api/v1/cart/discount", {
    method: "DELETE",
    schema: cartSchema,
  });
}

/**
 * Places the order with the policies the checkout showed: the API records their
 * version. The same idempotency key again returns the order the first attempt
 * placed, when its answer was lost.
 */
export async function checkoutCart(input: CheckoutInput, idempotencyKey: string): Promise<Order> {
  const body = checkoutInputSchema.parse(input);
  return browserRequest<Order>("/api/v1/orders/checkout", {
    method: "POST",
    schema: orderSchema,
    json: { ...body, terms_version: TERMS_VERSION },
    headers: { "Idempotency-Key": idempotencyKey },
  });
}

/** The PayPal page the customer approves the order's payment on. */
export async function startPayment(orderId: string): Promise<string> {
  const result = await browserRequest<PaymentStart>(`/api/v1/orders/${encodeURIComponent(orderId)}/payment`, {
    method: "POST",
    schema: paymentStartSchema,
  });
  return result.approve_url;
}

/** Takes the payment the customer approved on PayPal's page. */
export async function capturePayment(orderId: string): Promise<PaymentCapture> {
  return browserRequest<PaymentCapture>(`/api/v1/orders/${encodeURIComponent(orderId)}/payment/capture`, {
    method: "POST",
    schema: paymentCaptureSchema,
  });
}

export async function listOrders(
  input: OrdersListFilterInput,
): Promise<OrdersListResult> {
  const filter = ordersListFilterSchema.parse(input);
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  if (filter.status) {
    params.set("status", filter.status);
  }
  if (filter.from) {
    params.set("from", filter.from);
  }
  if (filter.to) {
    params.set("to", filter.to);
  }
  const result = await browserRequestWithMeta<z.infer<typeof orderSummarySchema>[]>(
    `/api/v1/orders?${params.toString()}`,
    { method: "GET", schema: z.array(orderSummarySchema) },
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

export async function getOrder(id: string): Promise<Order> {
  const parsedId = z.uuid().parse(id);
  return browserRequest<Order>(`/api/v1/orders/${parsedId}`, {
    method: "GET",
    schema: orderSchema,
  });
}

/** One bakery day's pickup slots and which are full; public, like the rules. */
export async function getPickupSlots(date: string): Promise<PickupSlots> {
  const day = z.iso.date().parse(date);
  return browserRequest<PickupSlots>(`/api/v1/store/pickup-slots?date=${day}`, {
    method: "GET",
    schema: pickupSlotsSchema,
  });
}

/** The pickup window checkout enforces; public, like the catalog. */
export async function getPickupRules(): Promise<PickupRules> {
  return browserRequest<PickupRules>("/api/v1/store/pickup-rules", {
    method: "GET",
    schema: pickupRulesSchema,
  });
}
