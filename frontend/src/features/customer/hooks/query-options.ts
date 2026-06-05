import type {
  CatalogCategoriesListFilterInput,
  CatalogProductsListFilterInput,
} from "../schemas";

export const customerQueryKeys = {
  catalogCategoriesRoot: ["customer", "catalog", "categories"] as const,
  catalogCategories: (filter: CatalogCategoriesListFilterInput) =>
    [...customerQueryKeys.catalogCategoriesRoot, filter] as const,
  catalogProductsRoot: ["customer", "catalog", "products"] as const,
  catalogProducts: (filter: CatalogProductsListFilterInput) =>
    [...customerQueryKeys.catalogProductsRoot, filter] as const,
};
