import type {
  CategoryListFilterInput,
  ComboListFilterInput,
  DiscountCodeListFilterInput,
  ProductListFilterInput,
} from "../schemas";

/** List queries: align with global default; explicit for manager catalog tables. */
export const MANAGER_LIST_STALE_TIME_MS = 30_000;

export const managerQueryKeys = {
  all: ["manager"] as const,
  categoriesRoot: ["manager", "categories"] as const,
  categories: (filter: CategoryListFilterInput) =>
    [...managerQueryKeys.categoriesRoot, filter] as const,
  category: (id: string) => [...managerQueryKeys.all, "category", id] as const,
  productsRoot: ["manager", "products"] as const,
  products: (filter: ProductListFilterInput) =>
    [...managerQueryKeys.productsRoot, filter] as const,
  product: (id: string) => [...managerQueryKeys.all, "product", id] as const,
  combosRoot: ["manager", "combos"] as const,
  combos: (filter: ComboListFilterInput) =>
    [...managerQueryKeys.combosRoot, filter] as const,
  combo: (id: string) => [...managerQueryKeys.all, "combo", id] as const,
  discountCodesRoot: ["manager", "discount-codes"] as const,
  discountCodes: (filter: DiscountCodeListFilterInput) =>
    [...managerQueryKeys.discountCodesRoot, filter] as const,
  discountCode: (id: string) =>
    [...managerQueryKeys.all, "discount-code", id] as const,
};
