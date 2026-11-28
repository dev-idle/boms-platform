import { ProductGridSkeleton } from "./product-grid-skeleton";

/** Initial catalog load — also the route's Suspense fallback, so both states look identical. */
export function ProductCatalogLoading() {
  return (
    <div aria-busy="true" className="storefront-container storefront-page">
      <div className="catalog-page__top">
        <h1 className="sr-only">All products</h1>
        <span className="skeleton block h-11 w-56 max-w-full" />
        <div className="catalog-chips">
          <div className="catalog-chips__list">
            {[0, 1, 2, 3, 4].map((chip) => (
              <span className="skeleton block h-11 w-32 shrink-0" key={chip} />
            ))}
          </div>
        </div>
      </div>
      <div className="catalog-page__body">
        <ProductGridSkeleton count={8} />
      </div>
    </div>
  );
}
