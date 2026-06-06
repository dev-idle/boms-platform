import type { StaffOrdersListFilterInput } from "../schemas";

export const staffQueryKeys = {
  root: ["staff"] as const,
  ordersRoot: ["staff", "orders"] as const,
  orders: (filter: StaffOrdersListFilterInput) =>
    [...staffQueryKeys.ordersRoot, filter] as const,
  order: (id: string) => [...staffQueryKeys.ordersRoot, id] as const,
};
