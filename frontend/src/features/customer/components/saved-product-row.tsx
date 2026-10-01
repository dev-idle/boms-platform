"use client";

import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useMoveToCart, useToggleSaved } from "../hooks";
import { SAVED_LIST_NOUN, type SavedProduct } from "../schemas";
import { AddToCartButton } from "./add-to-cart-button";

type SavedProductRowProps = {
  item: SavedProduct;
};

/**
 * One product on a list. A favorite goes in the cart and stays a favorite; a
 * wishlist product moves to the cart. A custom cake is configured first.
 */
export function SavedProductRow({ item }: SavedProductRowProps) {
  const toggle = useToggleSaved();
  const move = useMoveToCart();
  const { product } = item;

  let purchase;
  if (product.is_customizable) {
    purchase = (
      <Button asChild variant="outline">
        <Link href={ROUTE.productDetail(product.id)}>Customize</Link>
      </Button>
    );
  } else if (item.list === "wishlist") {
    purchase = (
      <Button
        aria-busy={move.isPending || undefined}
        disabled={move.isPending}
        type="button"
        onClick={() => move.mutate(product.id)}
      >
        {move.isPending ? "Moving…" : "Move to cart"}
      </Button>
    );
  } else {
    purchase = <AddToCartButton productId={product.id} />;
  }

  return (
    <li className="storefront-cart-line">
      <div className="storefront-cart-line__header">
        <div className="storefront-cart-line__copy">
          <Link className="storefront-cart-line__name storefront-inline-link" href={ROUTE.productDetail(product.id)}>
            {product.name}
          </Link>
          <p className="storefront-cart-line__meta text-caption">{product.category_name}</p>
          {product.sold_out_today ? <p className="catalog-sold-out">Sold out today</p> : null}
        </div>
        <p className="text-price storefront-cart-line__total">{formatPriceCents(product.price_cents)}</p>
      </div>
      <div className="storefront-cart-line__actions">
        {purchase}
        <button
          aria-busy={toggle.isPending || undefined}
          aria-label={`Remove ${product.name} from your ${SAVED_LIST_NOUN[item.list]}`}
          className="storefront-cart-remove"
          disabled={toggle.isPending}
          type="button"
          onClick={() => toggle.mutate({ list: item.list, product, onList: true })}
        >
          Remove
        </button>
      </div>
    </li>
  );
}
