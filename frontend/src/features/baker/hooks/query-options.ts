import type { BakerOrdersListFilterInput } from "../schemas";

export const bakerQueryKeys = {
  productionRoot: ["baker", "production"] as const,
  production: (filter: BakerOrdersListFilterInput) =>
    [...bakerQueryKeys.productionRoot, filter] as const,
  productionOrder: (id: string) => ["baker", "production", id] as const,
};
