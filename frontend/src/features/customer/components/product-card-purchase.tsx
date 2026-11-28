import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import type { CatalogProduct } from "@/lib/schemas/catalog";

import { ProductPurchaseActions } from "./product-purchase-actions";

type ProductCardPurchaseProps = {
  product: CatalogProduct;
};

/**
 * A product card's purchase: a ready product goes straight into the cart; a
 * custom cake opens its own page, where it is configured first.
 */
export function ProductCardPurchase({ product }: ProductCardPurchaseProps) {
  if (product.is_customizable) {
    return (
      <Button asChild size="sm" variant="outline">
        <Link href={ROUTE.productDetail(product.id)}>Customize</Link>
      </Button>
    );
  }

  return <ProductPurchaseActions product={product} secondary />;
}
