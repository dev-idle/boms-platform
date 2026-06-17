"use client";

import { useState } from "react";
import type { ReactNode } from "react";
import { z } from "zod";

import { StorefrontBrowseLink } from "@/components/layouts/storefront-browse-link";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { catalogProductImageUrl } from "@/lib/cloudinary/config";
import { isApiError } from "@/lib/errors";
import type { CatalogProduct } from "@/lib/schemas/catalog";
import { formatPriceCents } from "@/lib/validation/catalog";
import { cn } from "@/lib/utils";

import { useCatalogProduct } from "../hooks";
import { buildCatalogBrowseHref } from "../lib/catalog-browse-params";

type ProductDetailProps = {
  productId: string;
  /** Server-fetched product avoids a duplicate client request on first paint. */
  initialProduct?: CatalogProduct;
  purchaseActions?: ReactNode;
};

export function ProductDetail({
  productId,
  initialProduct,
  purchaseActions,
}: ProductDetailProps) {
  const [activeImageIndex, setActiveImageIndex] = useState(0);
  const isValidId = z.string().uuid().safeParse(productId).success;
  const productQuery = useCatalogProduct(productId, {
    enabled: isValidId && !initialProduct,
  });

  if (!isValidId) {
    return (
      <div className="catalog-detail storefront-container">
        <p className="text-sm text-muted">Invalid product link.</p>
      </div>
    );
  }

  const product = initialProduct ?? productQuery.data;

  if (!product && productQuery.isPending) {
    return (
      <div className="catalog-detail storefront-container">
        <InlineLoadingState />
      </div>
    );
  }

  if (!product && productQuery.isError) {
    const message =
      isApiError(productQuery.error) && productQuery.error.status === 404
        ? "Product not found."
        : "Failed to load product.";
    return (
      <div className="catalog-detail storefront-container">
        <p className="text-sm text-error">{message}</p>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="catalog-detail storefront-container">
        <p className="text-sm text-muted">Product not found.</p>
      </div>
    );
  }

  const imageUrls = product.image_urls;
  const heroIndex = Math.min(activeImageIndex, Math.max(imageUrls.length - 1, 0));
  const heroUrl = imageUrls[heroIndex]
    ? catalogProductImageUrl(imageUrls[heroIndex], 960)
    : undefined;

  return (
    <div className="catalog-detail storefront-container">
      <StorefrontBrowseLink
        href={buildCatalogBrowseHref({
          category: product.category_id,
          page: 1,
        })}
        withNavShell
      />

      <article className="catalog-detail__layout">
        <div className="catalog-detail__gallery">
          <div className="overflow-hidden rounded-card bg-mint shadow-rest">
            {heroUrl ? (
              // eslint-disable-next-line @next/next/no-img-element -- catalog URLs are external manager-provided links
              <img
                alt={product.name}
                className="aspect-square w-full object-cover"
                src={heroUrl}
              />
            ) : (
              <div className="flex aspect-square items-center justify-center text-caption">
                No image available
              </div>
            )}
          </div>
          {imageUrls.length > 1 ? (
            <div
              aria-label="Product image gallery"
              className="flex flex-wrap gap-2"
              role="list"
            >
              {imageUrls.map((url, index) => {
                const thumbUrl = catalogProductImageUrl(url, 160);
                if (!thumbUrl) {
                  return null;
                }
                return (
                  <button
                    aria-current={index === heroIndex ? "true" : undefined}
                    aria-label={`Show image ${index + 1} of ${imageUrls.length}`}
                    className={cn(
                      "overflow-hidden rounded-input border-2 bg-mint",
                      index === heroIndex ? "border-accent" : "border-transparent",
                    )}
                    key={`${url}-${index}`}
                    onClick={() => setActiveImageIndex(index)}
                    role="listitem"
                    type="button"
                  >
                    {/* eslint-disable-next-line @next/next/no-img-element -- catalog URLs are external manager-provided links */}
                    <img
                      alt=""
                      className="size-16 object-cover"
                      src={thumbUrl}
                    />
                  </button>
                );
              })}
            </div>
          ) : null}
        </div>

        <div className="catalog-detail__info">
          <p className="catalog-detail__category text-overline">
            {product.category_name}
          </p>
          <h1 className="catalog-detail__title">{product.name}</h1>
          <p className="catalog-detail__price text-price">
            {formatPriceCents(product.price_cents)}
          </p>
          {product.description ? (
            <p className="catalog-detail__description">{product.description}</p>
          ) : null}
          {purchaseActions ? (
            <div className="catalog-detail__actions">{purchaseActions}</div>
          ) : null}
        </div>
      </article>
    </div>
  );
}
