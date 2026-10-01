"use client";

import { ProductCatalog } from "@/features/catalog";
import { ProductSaveButtons } from "@/features/customer";

/** Composes the catalog with the customer's lists (FSD boundary at app layer). */
export function ProductCatalogSection() {
  return <ProductCatalog renderProductActions={(product) => <ProductSaveButtons product={product} />} />;
}
