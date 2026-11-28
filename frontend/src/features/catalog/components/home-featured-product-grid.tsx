import type { ReactNode } from "react";

import { StorefrontEmptyState } from "@/components/layouts/storefront-state";
import type { CatalogProduct } from "@/lib/schemas/catalog";

import { ProductCard } from "./product-card";

type HomeFeaturedProductGridProps = {
  products: CatalogProduct[];
  /** The save toggles and the purchase on each product's card, from the slice that owns them. */
  renderProductSaves?: (product: CatalogProduct) => ReactNode;
  renderProductPurchase?: (product: CatalogProduct) => ReactNode;
};

export function HomeFeaturedProductGrid({
  products,
  renderProductSaves,
  renderProductPurchase,
}: HomeFeaturedProductGridProps) {
  if (products.length === 0) {
    return (
      <div className="mt-12">
        <StorefrontEmptyState message="Please check back later." title="No products yet" />
      </div>
    );
  }

  return (
    <div className="catalog-product-grid catalog-product-grid--featured mt-12">
      {products.map((product) => (
        <ProductCard
          key={product.id}
          product={product}
          purchase={renderProductPurchase?.(product)}
          saves={renderProductSaves?.(product)}
        />
      ))}
    </div>
  );
}
