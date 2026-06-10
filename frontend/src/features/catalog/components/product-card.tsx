import Link from "next/link";

import { ROUTE } from "@/constants/routes";
import { formatPriceCents } from "@/lib/validation/catalog";
import { cn } from "@/lib/utils";

import type { CatalogProduct } from "@/lib/schemas/catalog";

type ProductCardProps = {
  product: CatalogProduct;
  featured?: boolean;
};

export function ProductCard({ product, featured = false }: ProductCardProps) {
  return (
    <article
      className={cn(
        "group overflow-hidden rounded-card bg-surface shadow-rest transition-[box-shadow,transform] duration-standard ease-default hover:shadow-hover",
        featured && "sm:flex sm:flex-row",
      )}
    >
      <Link
        className={cn("flex flex-col", featured && "sm:flex-1 sm:flex-row")}
        href={ROUTE.productDetail(product.id)}
      >
        <div
          className={cn(
            "relative overflow-hidden bg-blush",
            featured
              ? "aspect-square sm:aspect-auto sm:w-1/2"
              : "aspect-square",
          )}
        >
          {product.image_url ? (
            // eslint-disable-next-line @next/next/no-img-element -- catalog URLs are external manager-provided links
            <img
              alt={product.name}
              className="size-full rounded-t-card object-cover transition-transform duration-standard ease-default group-hover:scale-[1.02] motion-reduce:transform-none"
              src={product.image_url}
            />
          ) : (
            <div
              aria-hidden="true"
              className="flex size-full items-center justify-center text-caption"
            >
              No image
            </div>
          )}
        </div>
        <div className={cn("flex flex-col p-5", featured && "sm:justify-center sm:p-8")}>
          <p className="text-caption uppercase tracking-wide">
            {product.category_name}
          </p>
          <h2
            className={cn(
              "mt-1.5 font-heading font-medium leading-snug text-ink",
              featured ? "text-2xl" : "text-lg",
            )}
          >
            {product.name}
          </h2>
          <p className="text-price mt-3 text-base">
            {formatPriceCents(product.price_cents)}
          </p>
        </div>
      </Link>
    </article>
  );
}
