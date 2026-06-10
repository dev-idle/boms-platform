"use client";

import { ProductDetail } from "@/features/catalog/components/product-detail";
import { ProductPurchaseActions } from "@/features/customer";
import type { CatalogProduct } from "@/lib/schemas/catalog";

type ProductDetailPageProps = {
  productId: string;
  initialProduct: CatalogProduct;
};

/** Composes catalog display with customer purchase actions (FSD boundary at app layer). */
export function ProductDetailPage({
  productId,
  initialProduct,
}: ProductDetailPageProps) {
  return (
    <ProductDetail
      initialProduct={initialProduct}
      productId={productId}
      purchaseActions={
        <ProductPurchaseActions productId={initialProduct.id} />
      }
    />
  );
}
