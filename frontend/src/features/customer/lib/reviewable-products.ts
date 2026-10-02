import type { OrderStatus } from "@/lib/schemas/order";

type ReviewableOrder = {
  status: OrderStatus;
  items: ReadonlyArray<{ product_id?: string | null; name: string }>;
};

export type ReviewableProduct = {
  id: string;
  name: string;
};

/**
 * The products a customer may review on their order: once it is picked up,
 * each product line's product once, in the order's order. A combo is not one
 * of its products.
 */
export function reviewableProducts(order: ReviewableOrder): ReviewableProduct[] {
  if (order.status !== "fulfilled") {
    return [];
  }
  const products = new Map<string, string>();
  for (const item of order.items) {
    if (item.product_id && !products.has(item.product_id)) {
      products.set(item.product_id, item.name);
    }
  }
  return [...products].map(([id, name]) => ({ id, name }));
}
