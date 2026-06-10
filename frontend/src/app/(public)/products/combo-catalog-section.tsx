"use client";

import { ComboCatalog } from "@/features/catalog/components/combo-catalog";
import { ProductPurchaseActions } from "@/features/customer";

/** Composes catalog combos with customer cart actions (FSD boundary at app layer). */
export function ComboCatalogSection() {
  return (
    <ComboCatalog
      renderPurchaseActions={(comboId) => (
        <ProductPurchaseActions comboId={comboId} label="Add combo to cart" />
      )}
    />
  );
}
