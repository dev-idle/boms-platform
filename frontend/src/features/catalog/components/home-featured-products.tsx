import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { catalogProductFallbackImageUrl } from "@/constants/storefront-imagery";

import type { CatalogProduct } from "@/lib/schemas/catalog";
import { ProductCard } from "./product-card";

type HomeFeaturedProductsProps = {
  products: CatalogProduct[];
};

export function HomeFeaturedProducts({ products }: HomeFeaturedProductsProps) {
  return (
    <section
      aria-labelledby="home-featured-heading"
      className="storefront-section border-t border-border bg-bg"
    >
      <div className="storefront-container">
        <div className="text-center">
          <h2 className="text-h2" id="home-featured-heading">
            Fresh <span className="italic text-matcha-500">from the oven</span>
          </h2>
          <p className="mt-2 text-sm text-muted">
            Customer favorites and seasonal picks, made daily.
          </p>
        </div>

        {products.length === 0 ? (
          <p className="mt-8 text-center text-sm text-muted">
            Our shelves are being filled — check back soon or{" "}
            <Link
              className="font-medium text-matcha-500 underline-offset-4 hover:underline"
              href={ROUTE.products}
            >
              browse the shop
            </Link>
            .
          </p>
        ) : (
          <>
            <div className="catalog-product-grid mt-8">
              {products.map((product, index) => (
                <ProductCard
                  key={product.id}
                  fallbackImageUrl={catalogProductFallbackImageUrl(index)}
                  product={product}
                />
              ))}
            </div>
            <div className="mt-8 text-center">
              <Button asChild variant="outline">
                <Link href={ROUTE.products}>View all products</Link>
              </Button>
            </div>
          </>
        )}
      </div>
    </section>
  );
}
