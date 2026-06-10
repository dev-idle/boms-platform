import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { cn } from "@/lib/utils";

import type { CatalogProduct } from "@/lib/schemas/catalog";
import { ProductCard } from "./product-card";

type HomeFeaturedProductsProps = {
  products: CatalogProduct[];
};

export function HomeFeaturedProducts({ products }: HomeFeaturedProductsProps) {
  return (
    <section
      aria-labelledby="home-featured-heading"
      className="storefront-section bg-blush"
    >
      <div className="storefront-container">
        <div className="reveal flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2
              className="text-h2"
              id="home-featured-heading"
            >
              Fresh from the oven
            </h2>
            <p className="mt-3 text-sm text-ink-2">
              Customer favorites and seasonal picks, made daily.
            </p>
          </div>
          <Button asChild variant="outline">
            <Link href={ROUTE.products}>View all</Link>
          </Button>
        </div>

        {products.length === 0 ? (
          <p className="mt-10 text-sm text-ink-2">
            Our shelves are being filled — check back soon or{" "}
            <Link
              className="font-medium text-rose-500 underline-offset-4 hover:underline"
              href={ROUTE.products}
            >
              browse the shop
            </Link>
            .
          </p>
        ) : (
          <div className="reveal reveal-delay-1 mt-10 grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
            {products.map((product, index) => (
              <div
                key={product.id}
                className={cn(index === 0 && "sm:col-span-2")}
              >
                <ProductCard featured={index === 0} product={product} />
              </div>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
