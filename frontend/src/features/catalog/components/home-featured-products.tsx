import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";

import type { CatalogProduct } from "../schemas";
import { ProductCard } from "./product-card";

type HomeFeaturedProductsProps = {
  products: CatalogProduct[];
};

export function HomeFeaturedProducts({ products }: HomeFeaturedProductsProps) {
  return (
    <section
      aria-labelledby="home-featured-heading"
      className="border-t border-border bg-surface-alt/40 py-16 sm:py-20"
    >
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2
              className="font-heading text-2xl font-medium tracking-tight text-foreground sm:text-3xl"
              id="home-featured-heading"
            >
              Fresh from the oven
            </h2>
            <p className="mt-2 text-sm text-muted">
              Customer favorites and seasonal picks, made daily.
            </p>
          </div>
          <Button asChild variant="outline">
            <Link href={ROUTE.products}>View all</Link>
          </Button>
        </div>

        {products.length === 0 ? (
          <p className="mt-10 text-sm text-muted">
            Our shelves are being filled — check back soon or{" "}
            <Link
              className="font-medium text-foreground underline underline-offset-4"
              href={ROUTE.products}
            >
              browse the shop
            </Link>
            .
          </p>
        ) : (
          <div className="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
            {products.map((product) => (
              <ProductCard key={product.id} product={product} />
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
