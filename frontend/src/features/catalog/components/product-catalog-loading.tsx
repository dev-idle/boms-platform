import { InlineLoadingState } from "@/components/ui/loading-state";

/** Initial catalog load — also the route's Suspense fallback, so both states look identical. */
export function ProductCatalogLoading() {
  return (
    <div className="catalog-page">
      <h1 className="sr-only">Shop</h1>
      <div className="storefront-container catalog-page__loading-shell">
        <InlineLoadingState className="catalog-page__loading" />
      </div>
    </div>
  );
}
