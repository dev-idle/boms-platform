import type {
  CatalogCategoriesListFilterInput,
  CatalogCombosListFilterInput,
  CatalogProductsListFilterInput,
  OrdersListFilterInput,
} from "../schemas";

export const customerQueryKeys = {
  catalogCategoriesRoot: ["customer", "catalog", "categories"] as const,
  catalogCategories: (filter: CatalogCategoriesListFilterInput) =>
    [...customerQueryKeys.catalogCategoriesRoot, filter] as const,
  catalogProductsRoot: ["customer", "catalog", "products"] as const,
  catalogProducts: (filter: CatalogProductsListFilterInput) =>
    [...customerQueryKeys.catalogProductsRoot, filter] as const,
  catalogProduct: (id: string) => ["customer", "catalog", "product", id] as const,
  catalogCombosRoot: ["customer", "catalog", "combos"] as const,
  catalogCombos: (filter: CatalogCombosListFilterInput) =>
    [...customerQueryKeys.catalogCombosRoot, filter] as const,
  cart: ["customer", "cart"] as const,
  ordersRoot: ["customer", "orders"] as const,
  orders: (filter: OrdersListFilterInput) =>
    [...customerQueryKeys.ordersRoot, filter] as const,
  order: (id: string) => ["customer", "order", id] as const,
};
