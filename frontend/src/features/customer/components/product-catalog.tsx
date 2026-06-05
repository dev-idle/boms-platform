"use client";

import Link from "next/link";
import { useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { formatPriceCents } from "@/lib/validation/catalog";

import { ComboCatalog } from "./combo-catalog";

import { useCatalogCategories, useCatalogProducts } from "../hooks";

const PRODUCTS_PAGE_SIZE = 24;

export function ProductCatalog() {
  const [selectedCategoryId, setSelectedCategoryId] = useState<string | undefined>();
  const [page, setPage] = useState(1);

  const categoriesQuery = useCatalogCategories();
  const productsFilter = useMemo(
    () => ({
      page,
      page_size: PRODUCTS_PAGE_SIZE,
      category_id: selectedCategoryId,
    }),
    [page, selectedCategoryId],
  );
  const productsQuery = useCatalogProducts(productsFilter);

  const categories = categoriesQuery.data?.categories ?? [];
  const products = productsQuery.data?.products ?? [];
  const pagination = productsQuery.data?.pagination;

  function selectCategory(categoryId: string | undefined): void {
    setSelectedCategoryId(categoryId);
    setPage(1);
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
          Products
        </h1>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
          Browse our bakery catalog.
        </p>
      </div>

      {categoriesQuery.isPending ? (
        <p className="text-sm text-zinc-500">Loading categories…</p>
      ) : categoriesQuery.isError ? (
        <p className="text-sm text-red-600">Failed to load categories.</p>
      ) : (
        <div className="flex flex-wrap gap-2">
          <button
            className={`rounded-full px-4 py-2 text-sm font-medium ${
              selectedCategoryId === undefined
                ? "bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900"
                : "bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200"
            }`}
            onClick={() => selectCategory(undefined)}
            type="button"
          >
            All
          </button>
          {categories.map((category) => (
            <button
              key={category.id}
              className={`rounded-full px-4 py-2 text-sm font-medium ${
                selectedCategoryId === category.id
                  ? "bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900"
                  : "bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200"
              }`}
              onClick={() => selectCategory(category.id)}
              type="button"
            >
              {category.name}
            </button>
          ))}
        </div>
      )}

      {productsQuery.isPending ? (
        <p className="text-sm text-zinc-500">Loading products…</p>
      ) : productsQuery.isError ? (
        <p className="text-sm text-red-600">Failed to load products.</p>
      ) : products.length === 0 ? (
        <p className="text-sm text-zinc-500">No products available.</p>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {products.map((product) => (
            <Link
              key={product.id}
              className="rounded-lg border border-zinc-200 bg-white p-4 transition hover:border-zinc-400 dark:border-zinc-800 dark:bg-zinc-950 dark:hover:border-zinc-600"
              href={ROUTE.productDetail(product.id)}
            >
              <p className="text-xs font-medium uppercase tracking-wide text-zinc-500">
                {product.category_name}
              </p>
              <h2 className="mt-1 text-lg font-semibold text-zinc-900 dark:text-zinc-50">
                {product.name}
              </h2>
              {product.description ? (
                <p className="mt-2 line-clamp-2 text-sm text-zinc-600 dark:text-zinc-400">
                  {product.description}
                </p>
              ) : null}
              <p className="mt-3 text-base font-medium text-zinc-900 dark:text-zinc-50">
                {formatPriceCents(product.price_cents)}
              </p>
            </Link>
          ))}
        </div>
      )}

      {pagination && pagination.total_pages > 1 ? (
        <div className="flex items-center gap-2">
          <Button
            disabled={page <= 1}
            onClick={() => setPage((current) => Math.max(1, current - 1))}
            type="button"
            variant="outline"
          >
            Previous
          </Button>
          <span className="text-sm text-zinc-600">
            Page {pagination.page} of {pagination.total_pages}
          </span>
          <Button
            disabled={page >= pagination.total_pages}
            onClick={() => setPage((current) => current + 1)}
            type="button"
            variant="outline"
          >
            Next
          </Button>
        </div>
      ) : null}

      <ComboCatalog />
    </div>
  );
}
