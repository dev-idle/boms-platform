type HomeFeaturedProductGridSkeletonProps = {
  count: number;
};

export function HomeFeaturedProductGridSkeleton({ count }: HomeFeaturedProductGridSkeletonProps) {
  return (
    <div
      aria-busy="true"
      aria-label="Loading products"
      className="catalog-product-grid catalog-product-grid--featured mt-12"
    >
      {Array.from({ length: count }, (_, index) => (
        <div aria-hidden="true" className="catalog-product-card" key={index}>
          <div className="catalog-product-card__media skeleton" />
          <div className="catalog-product-card__body">
            <span className="skeleton block h-3 w-1/3" />
            <span className="skeleton block h-6 w-3/4" />
            <span className="skeleton block h-4 w-1/4" />
          </div>
        </div>
      ))}
    </div>
  );
}
