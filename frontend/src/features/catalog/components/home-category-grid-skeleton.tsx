type HomeCategoryGridSkeletonProps = {
  count: number;
};

export function HomeCategoryGridSkeleton({ count }: HomeCategoryGridSkeletonProps) {
  return (
    <ul aria-busy="true" aria-label="Loading categories" className="storefront-category-grid mt-10">
      {Array.from({ length: count }, (_, index) => (
        <li key={index}>
          <div aria-hidden="true" className="storefront-category-tile">
            <span className="skeleton h-3 w-6" />
            <span className="skeleton h-6 w-3/4" />
          </div>
        </li>
      ))}
    </ul>
  );
}
