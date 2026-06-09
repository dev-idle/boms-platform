import type { OrdersListFilterInput } from "../schemas";

export const customerQueryKeys = {
  cart: ["customer", "cart"] as const,
  ordersRoot: ["customer", "orders"] as const,
  orders: (filter: OrdersListFilterInput) =>
    [...customerQueryKeys.ordersRoot, filter] as const,
  order: (id: string) => ["customer", "order", id] as const,
};
