/**
 * Catalog feature — public storefront browse (guest + customer).
 *
 * Internal: api/, components/, hooks/, lib/, schemas/
 */
export {
  getCatalogProduct,
  listCatalogCategories,
  listCatalogCombos,
  listCatalogProducts,
} from "./api";
export {
  CatalogSearchForm,
  ComboCatalog,
  ProductCard,
  ProductCatalog,
  ProductDetail,
  StorefrontHome,
} from "./components";
export {
  catalogQueryKeys,
  useCatalogBrowseFilters,
  useCatalogCategories,
  useCatalogCombos,
  useCatalogProduct,
  useCatalogProducts,
} from "./hooks";
export {
  buildCatalogBrowseHref,
  parseCatalogBrowseParams,
  toCatalogProductsFilter,
} from "./lib/catalog-browse-params";
export type { CatalogBrowseParams } from "./lib/catalog-browse-params";
export {
  catalogCategoriesListFilterSchema,
  catalogCategorySchema,
  catalogComboSchema,
  catalogCombosListFilterSchema,
  catalogProductSchema,
  catalogProductsListFilterSchema,
} from "./schemas";
export type {
  CatalogCategoriesListFilterInput,
  CatalogCategoriesListResult,
  CatalogCategory,
  CatalogCombo,
  CatalogCombosListFilterInput,
  CatalogCombosListResult,
  CatalogProduct,
  CatalogProductsListFilterInput,
  CatalogProductsListResult,
} from "./schemas";
