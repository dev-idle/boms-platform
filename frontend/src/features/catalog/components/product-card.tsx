import Link from "next/link";
import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";
import { formatPriceCents } from "@/lib/validation/catalog";

import type { CatalogProduct } from "@/lib/schemas/catalog";

import { CatalogTile } from "./catalog-tile";

type ProductCardProps = {
  product: CatalogProduct;
  /** Hide category label when the browse view already filters by category. */
  showCategory?: boolean;
  /** The save toggles, at the end of the price row. */
  saves?: ReactNode;
  /** The purchase action that closes the card. */
  purchase?: ReactNode;
};

/**
 * A product in a grid: the photo and the name link to its page; the price row
 * and the purchase sit outside that link, so each control is its own target.
 */
export function ProductCard({ product, showCategory = true, saves, purchase }: ProductCardProps) {
  return (
    <article className="catalog-product-card">
      <Link className="catalog-product-card__link" href={ROUTE.productDetail(product.id)}>
        <CatalogTile name={product.name} scale="card" url={product.image_urls[0]} />
        {showCategory ? <p className="catalog-product-card__eyebrow">{product.category_name}</p> : null}
        <h3 className="catalog-product-card__name">{product.name}</h3>
        {product.sold_out_today ? <p className="catalog-sold-out">Sold out today</p> : null}
      </Link>
      <div className="catalog-product-card__footer">
        <p className="catalog-product-card__price text-price">{formatPriceCents(product.price_cents)}</p>
        {saves}
      </div>
      {purchase ? <div className="catalog-product-card__purchase">{purchase}</div> : null}
    </article>
  );
}
