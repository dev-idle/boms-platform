import { cn } from "@/lib/utils";

type ProductGridSkeletonProps = {
  count: number;
  className?: string;
};

/** A product grid's cards while they load, in the card's own shape. */
export function ProductGridSkeleton({ count, className }: ProductGridSkeletonProps) {
  return (
    <div aria-busy="true" aria-label="Loading products" className={cn("catalog-product-grid", className)}>
      {Array.from({ length: count }, (_, index) => (
        <div aria-hidden="true" className="catalog-product-card" key={index}>
          <div className="catalog-tile catalog-tile--card skeleton" />
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
