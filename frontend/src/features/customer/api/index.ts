import { z } from "zod";

import {
  browserRequest,
  browserRequestWithMeta,
} from "@/lib/browser-api-client";
import { parsePaginatedList } from "@/lib/pagination/parse-paginated-list";

import {
  addCartItemInputSchema,
  applyCartDiscountInputSchema,
  cartSchema,
  catalogCategoriesListFilterSchema,
  catalogCategorySchema,
  catalogComboSchema,
  catalogCombosListFilterSchema,
  catalogProductSchema,
  catalogProductsListFilterSchema,
  orderSchema,
  orderSummarySchema,
  ordersListFilterSchema,
  updateCartItemInputSchema,
  type AddCartItemInput,
  type ApplyCartDiscountInput,
  type Cart,
  type CatalogCategoriesListFilterInput,
  type CatalogCategoriesListResult,
  type CatalogCategory,
  type CatalogCombo,
  type CatalogCombosListFilterInput,
  type CatalogCombosListResult,
  type CatalogProduct,
  type CatalogProductsListFilterInput,
  type CatalogProductsListResult,
  type Order,
  type OrdersListFilterInput,
  type OrdersListResult,
  type UpdateCartItemInput,
} from "../schemas";

export async function listCatalogCategories(
  input: CatalogCategoriesListFilterInput = { page: 1, page_size: 100 },
): Promise<CatalogCategoriesListResult> {
  const filter = catalogCategoriesListFilterSchema.parse(input);
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  const result = await browserRequestWithMeta<CatalogCategory[]>(
    `/api/v1/catalog/categories?${params.toString()}`,
    { method: "GET", schema: z.array(catalogCategorySchema) },
  );
  const parsed = parsePaginatedList(result.data, result.meta, {
    page: filter.page,
    page_size: filter.page_size,
  });
  return {
    categories: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}

export async function listCatalogProducts(
  input: CatalogProductsListFilterInput,
): Promise<CatalogProductsListResult> {
  const filter = catalogProductsListFilterSchema.parse(input);
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  if (filter.category_id) {
    params.set("category_id", filter.category_id);
  }
  if (filter.search) {
    params.set("search", filter.search);
  }
  const result = await browserRequestWithMeta<CatalogProduct[]>(
    `/api/v1/catalog/products?${params.toString()}`,
    { method: "GET", schema: z.array(catalogProductSchema) },
  );
  const parsed = parsePaginatedList(result.data, result.meta, {
    page: filter.page,
    page_size: filter.page_size,
  });
  return {
    products: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}

export async function getCatalogProduct(id: string): Promise<CatalogProduct> {
  const parsedId = z.string().uuid().parse(id);
  return browserRequest<CatalogProduct>(`/api/v1/catalog/products/${parsedId}`, {
    method: "GET",
    schema: catalogProductSchema,
  });
}

export async function listCatalogCombos(
  input: CatalogCombosListFilterInput,
): Promise<CatalogCombosListResult> {
  const filter = catalogCombosListFilterSchema.parse(input);
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  const result = await browserRequestWithMeta<CatalogCombo[]>(
    `/api/v1/catalog/combos?${params.toString()}`,
    { method: "GET", schema: z.array(catalogComboSchema) },
  );
  const parsed = parsePaginatedList(result.data, result.meta, {
    page: filter.page,
    page_size: filter.page_size,
  });
  return {
    combos: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}

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
  const id = z.string().uuid().parse(itemId);
  const body = updateCartItemInputSchema.parse(input);
  return browserRequest<Cart>(`/api/v1/cart/items/${id}`, {
    method: "PATCH",
    schema: cartSchema,
    json: body,
  });
}

export async function removeCartItem(itemId: string): Promise<Cart> {
  const id = z.string().uuid().parse(itemId);
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

export async function checkoutCart(): Promise<Order> {
  return browserRequest<Order>("/api/v1/orders/checkout", {
    method: "POST",
    schema: orderSchema,
  });
}

export async function listOrders(
  input: OrdersListFilterInput,
): Promise<OrdersListResult> {
  const filter = ordersListFilterSchema.parse(input);
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
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
  const parsedId = z.string().uuid().parse(id);
  return browserRequest<Order>(`/api/v1/orders/${parsedId}`, {
    method: "GET",
    schema: orderSchema,
  });
}

