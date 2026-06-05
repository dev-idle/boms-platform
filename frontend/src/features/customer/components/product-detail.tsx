"use client";

import Link from "next/link";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { formatPriceCents } from "@/lib/validation/catalog";

import { AddToCartButton } from "./add-to-cart-button";
import { useCatalogProduct } from "../hooks";

type ProductDetailProps = {
  productId: string;
};

export function ProductDetail({ productId }: ProductDetailProps) {
  const isValidId = z.string().uuid().safeParse(productId).success;
  const productQuery = useCatalogProduct(productId);

  if (!isValidId) {
    return <p className="text-sm text-zinc-500">Invalid product link.</p>;
  }

  if (productQuery.isPending) {
    return <p className="text-sm text-zinc-500">Loading product…</p>;
  }

  if (productQuery.isError) {
    const message =
      isApiError(productQuery.error) && productQuery.error.status === 404
        ? "Product not found."
        : "Failed to load product.";
    return <p className="text-sm text-red-600">{message}</p>;
  }

  const product = productQuery.data;
  if (!product) {
    return <p className="text-sm text-zinc-500">Product not found.</p>;
  }

  return (
    <div className="space-y-6">
      <Link href={ROUTE.products}>
        <Button type="button" variant="outline">
          Back to products
        </Button>
      </Link>

      <article className="rounded-lg border border-zinc-200 bg-white p-6 dark:border-zinc-800 dark:bg-zinc-950">
        <p className="text-xs font-medium uppercase tracking-wide text-zinc-500">
          {product.category_name}
        </p>
        <h1 className="mt-2 text-3xl font-semibold text-zinc-900 dark:text-zinc-50">
          {product.name}
        </h1>
        {product.image_url ? (
          // eslint-disable-next-line @next/next/no-img-element -- catalog URLs are external manager-provided links
          <img
            alt={product.name}
            className="mt-6 max-h-80 w-full rounded-lg object-cover"
            src={product.image_url}
          />
        ) : null}
        {product.description ? (
          <p className="mt-6 text-base text-zinc-600 dark:text-zinc-300">
            {product.description}
          </p>
        ) : null}
        <p className="mt-6 text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
          {formatPriceCents(product.price_cents)}
        </p>
        <div className="mt-6">
          <AddToCartButton productId={product.id} />
        </div>
      </article>
    </div>
  );
}
