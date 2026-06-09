import Link from "next/link";

import { ROUTE } from "@/constants/routes";
import { formatPriceCents } from "@/lib/validation/catalog";

import type { CatalogProduct } from "../schemas";

type ProductCardProps = {
  product: CatalogProduct;
};

export function ProductCard({ product }: ProductCardProps) {
  return (
    <article className="group overflow-hidden rounded-lg border border-border bg-surface transition-colors duration-default ease-default hover:border-border-strong">
      <Link
        className="flex flex-col"
        href={ROUTE.productDetail(product.id)}
      >
        <div className="relative aspect-square overflow-hidden bg-surface-alt">
          {product.image_url ? (
            // eslint-disable-next-line @next/next/no-img-element -- catalog URLs are external manager-provided links
            <img
              alt={product.name}
              className="size-full object-cover transition-transform duration-default ease-default group-hover:scale-[1.02]"
              src={product.image_url}
            />
          ) : (
            <div
              aria-hidden="true"
              className="flex size-full items-center justify-center text-sm text-subtle"
            >
              No image
            </div>
          )}
        </div>
        <div className="flex flex-col p-4">
          <p className="text-xs font-medium uppercase tracking-wide text-subtle">
            {product.category_name}
          </p>
          <h2 className="mt-1 font-heading text-lg font-medium leading-snug text-foreground">
            {product.name}
          </h2>
          <p className="mt-2 text-base font-medium text-foreground">
            {formatPriceCents(product.price_cents)}
          </p>
        </div>
      </Link>
    </article>
  );
}
