"use client";

import Link from "next/link";
import type { ReactNode } from "react";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { catalogProductImageUrl } from "@/lib/cloudinary/config";
import { isApiError } from "@/lib/errors";
import type { CatalogProduct } from "@/lib/schemas/catalog";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useCatalogProduct } from "../hooks";

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
  const isValidId = z.string().uuid().safeParse(productId).success;
  const productQuery = useCatalogProduct(productId, {
    enabled: isValidId && !initialProduct,
  });

  if (!isValidId) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
        <p className="text-sm text-muted">Invalid product link.</p>
      </div>
    );
  }

  const product = initialProduct ?? productQuery.data;

  if (!product && productQuery.isPending) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
        <p className="text-sm text-muted">Loading product…</p>
      </div>
    );
  }

  if (!product && productQuery.isError) {
    const message =
      isApiError(productQuery.error) && productQuery.error.status === 404
        ? "Product not found."
        : "Failed to load product.";
    return (
      <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
        <p className="text-sm text-error">{message}</p>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
        <p className="text-sm text-muted">Product not found.</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
      <Link href={ROUTE.products}>
        <Button type="button" variant="ghost">
          Back to shop
        </Button>
      </Link>

      <article className="mt-8 grid gap-10 lg:grid-cols-2 lg:gap-16">
        <div className="overflow-hidden rounded-card bg-mint shadow-rest">
          {product.image_url ? (
            // eslint-disable-next-line @next/next/no-img-element -- catalog URLs are external manager-provided links
            <img
              alt={product.name}
              className="aspect-square w-full object-cover"
              src={catalogProductImageUrl(product.image_url, 960)}
            />
          ) : (
            <div className="flex aspect-square items-center justify-center text-caption">
              No image available
            </div>
          )}
        </div>

        <div>
          <p className="text-caption uppercase tracking-wide">
            {product.category_name}
          </p>
          <h1 className="mt-2">
            {product.name}
          </h1>
          <p className="text-price mt-4 text-2xl">
            {formatPriceCents(product.price_cents)}
          </p>
          {product.description ? (
            <p className="mt-6 text-base leading-relaxed text-muted">
              {product.description}
            </p>
          ) : null}
          {purchaseActions ? <div className="mt-8">{purchaseActions}</div> : null}
        </div>
      </article>
    </div>
  );
}
