"use client";

import Link from "next/link";

import { CloseIcon } from "@/components/icons/storefront-icons";
import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { CatalogTile } from "@/features/catalog";
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
      <Button asChild size="sm" variant="outline">
        <Link href={ROUTE.productDetail(product.id)}>Customize</Link>
      </Button>
    );
  } else if (item.list === "wishlist") {
    purchase = (
      <Button
        aria-busy={move.isPending || undefined}
        disabled={move.isPending}
        onClick={() => move.mutate(product.id)}
        size="sm"
        type="button"
        variant="outline"
      >
        {move.isPending ? "Moving…" : "Move to cart"}
      </Button>
    );
  } else {
    purchase = <AddToCartButton productId={product.id} secondary />;
  }

  return (
    <li className="storefront-saved-line">
      <Link aria-hidden="true" className="storefront-saved-line__tile" href={ROUTE.productDetail(product.id)} tabIndex={-1}>
        <CatalogTile name={product.name} scale="row" url={product.image_urls[0]} />
      </Link>
      <div className="storefront-saved-line__copy">
        <p className="storefront-line-eyebrow">{product.category_name}</p>
        <Link className="storefront-line-name" href={ROUTE.productDetail(product.id)}>
          {product.name}
        </Link>
        <p className="storefront-line-meta">
          <span className="text-price">{formatPriceCents(product.price_cents)}</span>
        </p>
        {product.sold_out_today ? <p className="catalog-sold-out">Sold out today</p> : null}
      </div>
      <div className="storefront-saved-line__actions">
        {purchase}
        <button
          aria-busy={toggle.isPending || undefined}
          aria-label={`Remove ${product.name} from your ${SAVED_LIST_NOUN[item.list]}`}
          className="storefront-line-remove"
          disabled={toggle.isPending}
          onClick={() => toggle.mutate({ list: item.list, product, onList: true })}
          title="Remove"
          type="button"
        >
          <CloseIcon />
        </button>
      </div>
    </li>
  );
}
