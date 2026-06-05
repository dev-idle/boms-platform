/**
 * Customer feature — storefront browse (customer-only).
 *
 * Internal: api/, components/, hooks/, schemas/
 */
export { listCatalogCategories, listCatalogProducts } from "./api";
export { ProductCatalog } from "./components";
export {
  customerQueryKeys,
  useCatalogCategories,
  useCatalogProducts,
} from "./hooks";
export {
  catalogCategoriesListFilterSchema,
  catalogCategorySchema,
  catalogProductSchema,
  catalogProductsListFilterSchema,
} from "./schemas";
export type {
  CatalogCategoriesListFilterInput,
  CatalogCategoriesListResult,
  CatalogCategory,
  CatalogProduct,
  CatalogProductsListFilterInput,
  CatalogProductsListResult,
} from "./schemas";
