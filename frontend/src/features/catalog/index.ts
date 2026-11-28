/**
 * Catalog feature — public storefront browse (guest + customer).
 * The only import path for app/ routes, layouts and other slices.
 */
export {
  CatalogPagination,
  CatalogTile,
  ComboCatalog,
  HomeCategories,
  HomeCategoryGrid,
  HomeCategoryGridSkeleton,
  HomeCta,
  HomeFeaturedProductGrid,
  HomeFeaturedProducts,
  HomeHero,
  HomeSectionError,
  HomeUspStrip,
  ProductCatalog,
  ProductCatalogLoading,
  ProductDetail,
  ProductGridSkeleton,
} from "./components";
export { useCatalogCombos, useCatalogProducts, useCombosOnSale } from "./hooks";
export {
  buildCatalogBrowseHref,
  CATALOG_SEARCH_MAX_LENGTH,
} from "./lib/catalog-browse-params";
export { CATALOG_COMBOS_HEADING_ID } from "./lib/scroll-to-storefront-anchor";
export { CUSTOM_CAKE_BROWSE_HREF } from "./lib/storefront-links";
export { catalogItemInitial } from "./lib/catalog-item-initial";
