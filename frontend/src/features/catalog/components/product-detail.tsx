"use client";

import Link from "next/link";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { ProductPurchaseActions } from "@/features/customer";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useCatalogProduct } from "../hooks";

type ProductDetailProps = {
  productId: string;
};

export function ProductDetail({ productId }: ProductDetailProps) {
  const isValidId = z.string().uuid().safeParse(productId).success;
  const productQuery = useCatalogProduct(productId);

  if (!isValidId) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
        <p className="text-sm text-muted">Invalid product link.</p>
      </div>
    );
  }

  if (productQuery.isPending) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
        <p className="text-sm text-muted">Loading product…</p>
      </div>
    );
  }

  if (productQuery.isError) {
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

  const product = productQuery.data;
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
        <div className="overflow-hidden rounded-lg border border-border bg-surface-alt">
          {product.image_url ? (
            // eslint-disable-next-line @next/next/no-img-element -- catalog URLs are external manager-provided links
            <img
              alt={product.name}
              className="aspect-square w-full object-cover"
              src={product.image_url}
            />
          ) : (
            <div className="flex aspect-square items-center justify-center text-sm text-subtle">
              No image available
            </div>
          )}
        </div>

        <div>
          <p className="text-xs font-medium uppercase tracking-wide text-subtle">
            {product.category_name}
          </p>
          <h1 className="mt-2 font-heading text-3xl font-medium tracking-tight text-foreground sm:text-4xl">
            {product.name}
          </h1>
          <p className="mt-4 text-2xl font-medium text-foreground">
            {formatPriceCents(product.price_cents)}
          </p>
          {product.description ? (
            <p className="mt-6 text-base leading-relaxed text-muted">
              {product.description}
            </p>
          ) : null}
          <div className="mt-8">
            <ProductPurchaseActions productId={product.id} />
          </div>
        </div>
      </article>
    </div>
  );
}
