"use client";

import Link from "next/link";
import { useState } from "react";
import type { ReactNode } from "react";
import { z } from "zod";

import { SHOP_BACK, StorefrontBackLink } from "@/components/layouts/storefront-back-link";
import { StorefrontPageHeader } from "@/components/layouts/storefront-page-header";
import { StorefrontEmptyState, StorefrontErrorState } from "@/components/layouts/storefront-state";
import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { catalogProductImageUrl } from "@/lib/cloudinary/config";
import { isApiError } from "@/lib/errors";
import type { CatalogProduct } from "@/lib/schemas/catalog";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useCatalogProduct } from "../hooks";
import { buildCatalogBrowseHref } from "../lib/catalog-browse-params";
import { CatalogTile } from "./catalog-tile";
import { ProductReviews } from "./product-reviews";

/** The page's name while there is no product to name. */
const PRODUCT_PAGE_TITLE = "Product";

type ProductDetailProps = {
  productId: string;
  /** Server-fetched product avoids a duplicate client request on first paint. */
  initialProduct?: CatalogProduct;
  purchaseActions?: ReactNode;
  /** What the options chosen in the purchase actions add to the price. */
  optionsCents?: number;
};

/** A product the shop does not have: said once, with the way back to the shop. */
function ProductMissing() {
  return (
    <div className="storefront-container storefront-page storefront-page-stack">
      <StorefrontPageHeader back={SHOP_BACK} title={PRODUCT_PAGE_TITLE} />
      <StorefrontEmptyState
        action={
          <Button asChild>
            <Link href={ROUTE.products}>Browse all products</Link>
          </Button>
        }
        message="It may no longer be on sale."
        title="This product is not in the shop"
      />
    </div>
  );
}

/** The page's own shape while the product loads: the photo, then its details. */
function ProductDetailSkeleton() {
  return (
    <div aria-busy="true" className="storefront-container storefront-page">
      <span className="skeleton block h-5 w-40" />
      <div className="catalog-detail__layout">
        <span className="catalog-tile catalog-tile--hero skeleton" />
        <div className="catalog-detail__info grid content-start gap-5">
          <span className="skeleton block h-12 w-4/5" />
          <span className="skeleton block h-8 w-24" />
          <span className="skeleton block h-20 w-full" />
          <span className="skeleton block h-12 w-44" />
        </div>
      </div>
    </div>
  );
}

export function ProductDetail({
  productId,
  initialProduct,
  purchaseActions,
  optionsCents = 0,
}: ProductDetailProps) {
  const [activeImageIndex, setActiveImageIndex] = useState(0);
  const isValidId = z.uuid().safeParse(productId).success;
  const productQuery = useCatalogProduct(productId, {
    enabled: isValidId && !initialProduct,
  });

  if (!isValidId) {
    return <ProductMissing />;
  }

  const product = initialProduct ?? productQuery.data;

  if (!product && productQuery.isPending) {
    return <ProductDetailSkeleton />;
  }

  if (!product) {
    if (isApiError(productQuery.error) && productQuery.error.status === 404) {
      return <ProductMissing />;
    }
    return (
      <div className="storefront-container storefront-page storefront-page-stack">
        <StorefrontPageHeader back={SHOP_BACK} title={PRODUCT_PAGE_TITLE} />
        <StorefrontErrorState
          message="Check your connection and try again."
          onRetry={() => void productQuery.refetch()}
          title="We could not load this product"
        />
      </div>
    );
  }

  const imageUrls = product.image_urls;
  const heroIndex = Math.min(activeImageIndex, Math.max(imageUrls.length - 1, 0));

  return (
    <div className="storefront-container storefront-page">
      <StorefrontBackLink
        href={buildCatalogBrowseHref({ category: product.category_id, page: 1 })}
        label={product.category_name}
      />

      <article className="catalog-detail__layout">
        <div className="catalog-detail__gallery">
          <CatalogTile alt={product.name} name={product.name} scale="hero" url={imageUrls[heroIndex]} />
          {imageUrls.length > 1 ? (
            <div aria-label="Product photos" className="catalog-detail__thumbs" role="group">
              {imageUrls.map((url, index) => (
                <button
                  aria-current={index === heroIndex ? "true" : undefined}
                  aria-label={`Show photo ${index + 1} of ${imageUrls.length}`}
                  className="catalog-detail__thumb"
                  key={`${url}-${index}`}
                  onClick={() => setActiveImageIndex(index)}
                  type="button"
                >
                  {/* eslint-disable-next-line @next/next/no-img-element -- manager-provided catalog URL */}
                  <img alt="" src={catalogProductImageUrl(url, 128)} />
                </button>
              ))}
            </div>
          ) : null}
        </div>

        <div className="catalog-detail__info">
          <h1 className="catalog-detail__title">{product.name}</h1>
          <p aria-live="polite" className="catalog-detail__price text-price">
            {formatPriceCents(product.price_cents + optionsCents)}
          </p>
          {product.sold_out_today ? (
            <p className="catalog-detail__sold-out">
              <span className="catalog-sold-out">Sold out today</span>
              <span className="text-caption">You can still order it for a later day.</span>
            </p>
          ) : null}
          {product.description ? <p className="catalog-detail__description">{product.description}</p> : null}
          {purchaseActions ? <div className="catalog-detail__actions">{purchaseActions}</div> : null}
        </div>
      </article>

      <ProductReviews productId={product.id} />
    </div>
  );
}
