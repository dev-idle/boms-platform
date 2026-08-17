import Link from "next/link";

import { ROUTE } from "@/constants/routes";
import { primaryCatalogProductImageUrl } from "@/lib/cloudinary/config";
import { formatPriceCents } from "@/lib/validation/catalog";

import type { CatalogProduct } from "@/lib/schemas/catalog";

type ProductCardProps = {
  product: CatalogProduct;
  /** Shown when the catalog item has no manager-provided image URL. */
  fallbackImageUrl?: string;
  /** Hide category label when the browse view already filters by category. */
  showCategory?: boolean;
};

export function ProductCard({
  product,
  fallbackImageUrl,
  showCategory = true,
}: ProductCardProps) {
  const imageSources =
    product.image_urls.length > 0
      ? product.image_urls
      : fallbackImageUrl
        ? [fallbackImageUrl]
        : [];
  const imageUrl = primaryCatalogProductImageUrl(imageSources, 720);

  return (
    <article className="catalog-product-card group">
      <Link
        className="catalog-product-card__link"
        href={ROUTE.productDetail(product.id)}
      >
        <div className="catalog-product-card__media">
          {imageUrl ? (
            // eslint-disable-next-line @next/next/no-img-element -- catalog URLs are external manager-provided links
            <img
              alt={product.name}
              className="catalog-product-card__image"
              src={imageUrl}
            />
          ) : (
            <div aria-hidden="true" className="catalog-product-card__placeholder">
              No image
            </div>
          )}
        </div>
        <div className="catalog-product-card__body">
          {showCategory ? (
            <p className="catalog-product-card__eyebrow">
              {product.category_name}
            </p>
          ) : null}
          <h2 className="catalog-product-card__name">{product.name}</h2>
          <p className="catalog-product-card__price text-price">
            {formatPriceCents(product.price_cents)}
          </p>
        </div>
      </Link>
    </article>
  );
}
