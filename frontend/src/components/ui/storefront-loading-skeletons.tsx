import { LOADING_MESSAGE } from "@/constants/loading-copy";

const PRODUCT_SKELETON_COUNT = 4;
const CATEGORY_PILL_COUNT = 3;

type StorefrontProductGridSkeletonProps = {
  count?: number;
};

/** Static layout placeholders — one shared shimmer on the grid container (perf). */
export function StorefrontProductGridSkeleton({
  count = PRODUCT_SKELETON_COUNT,
}: StorefrontProductGridSkeletonProps) {
  return (
    <div
      aria-busy="true"
      aria-label={LOADING_MESSAGE}
      className="storefront-product-grid-skeleton"
      role="status"
    >
      {Array.from({ length: count }, (_, index) => (
        <div className="storefront-product-card-skeleton" key={index}>
          <div className="storefront-product-card-skeleton__image" />
          <div className="storefront-product-card-skeleton__body">
            <div className="storefront-product-card-skeleton__title" />
            <div className="storefront-product-card-skeleton__price" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function StorefrontCategoryPillsSkeleton() {
  return (
    <div
      aria-busy="true"
      aria-label={LOADING_MESSAGE}
      className="storefront-category-pills-skeleton"
      role="status"
    >
      {Array.from({ length: CATEGORY_PILL_COUNT }, (_, index) => (
        <div
          className="storefront-category-pill-skeleton"
          key={index}
        />
      ))}
    </div>
  );
}
