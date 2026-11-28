"use client";

import { ProductCatalog } from "@/features/catalog";
import { ProductCardPurchase, ProductSaveButtons } from "@/features/customer";

import { ComboCatalogSection } from "./combo-catalog-section";

/** Composes the catalog with the customer's lists and cart (FSD boundary at app layer). */
export function ProductCatalogSection() {
  return (
    <ProductCatalog
      combos={<ComboCatalogSection />}
      renderProductPurchase={(product) => <ProductCardPurchase product={product} />}
      renderProductSaves={(product) => <ProductSaveButtons product={product} />}
    />
  );
}
