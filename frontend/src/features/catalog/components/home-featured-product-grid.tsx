import Link from "next/link";
import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";
import { catalogProductFallbackImageUrl } from "@/constants/storefront-imagery";
import type { CatalogProduct } from "@/lib/schemas/catalog";

import { ProductCard } from "./product-card";

type HomeFeaturedProductGridProps = {
  products: CatalogProduct[];
  /** What each product's card offers a visitor, from the slice that owns it. */
  renderProductActions?: (product: CatalogProduct) => ReactNode;
};

export function HomeFeaturedProductGrid({ products, renderProductActions }: HomeFeaturedProductGridProps) {
  if (products.length === 0) {
    return (
      <p className="mt-8 text-sm text-muted">
        Our shelves are being filled — check back soon or{" "}
        <Link className="storefront-inline-link" href={ROUTE.products}>
          browse the shop
        </Link>
        .
      </p>
    );
  }

  return (
    <div className="catalog-product-grid catalog-product-grid--featured mt-12">
      {products.map((product, index) => (
        <ProductCard
          key={product.id}
          actions={renderProductActions?.(product)}
          fallbackImageUrl={catalogProductFallbackImageUrl(index)}
          product={product}
        />
      ))}
    </div>
  );
}
