"use client";

import { useState } from "react";

import { ProductDetail } from "@/features/catalog";
import { ProductPurchaseActions, ProductSaveButtons } from "@/features/customer";
import type { CatalogProduct } from "@/lib/schemas/catalog";

type ProductDetailPageProps = {
  productId: string;
  initialProduct: CatalogProduct;
};

/**
 * Composes catalog display with customer purchase actions and lists (FSD boundary at app layer):
 * the price the catalog shows follows the options chosen in the customer's form.
 */
export function ProductDetailPage({
  productId,
  initialProduct,
}: ProductDetailPageProps) {
  const [optionsCents, setOptionsCents] = useState(0);

  return (
    <ProductDetail
      initialProduct={initialProduct}
      optionsCents={optionsCents}
      productId={productId}
      purchaseActions={
        <ProductPurchaseActions product={initialProduct} onOptionsCentsChange={setOptionsCents}>
          <ProductSaveButtons product={initialProduct} />
        </ProductPurchaseActions>
      }
    />
  );
}
