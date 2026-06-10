/**
 * Catalog feature — public storefront browse (guest + customer).
 *
 * Prefer deep imports from app/ RSC routes (`components/`, `hooks/query-options`).
 * This barrel is for client-side feature consumers only.
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
  catalogCategoriesQueryOptions,
  catalogCombosQueryOptions,
  catalogProductQueryOptions,
  catalogProductsQueryOptions,
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
