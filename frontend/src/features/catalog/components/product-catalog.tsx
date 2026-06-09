"use client";

import { useMemo } from "react";

import { Button } from "@/components/ui/button";

import { toCatalogProductsFilter } from "../lib/catalog-browse-params";
import {
  useCatalogBrowseFilters,
  useCatalogCategories,
  useCatalogProducts,
} from "../hooks";
import { CatalogPagination } from "./catalog-pagination";
import { CatalogSearchForm } from "./catalog-search-form";
import { ProductCard } from "./product-card";

const PRODUCTS_PAGE_SIZE = 24;

export function ProductCatalog() {
  const {
    params,
    setSearch,
    setCategory,
    setPage,
    clearFilters,
    hasActiveFilters,
  } = useCatalogBrowseFilters();

  const productsFilter = useMemo(
    () => toCatalogProductsFilter(params, PRODUCTS_PAGE_SIZE),
    [params],
  );

  const categoriesQuery = useCatalogCategories();
  const productsQuery = useCatalogProducts(productsFilter);

  const categories = categoriesQuery.data?.categories ?? [];
  const products = productsQuery.data?.products ?? [];
  const pagination = productsQuery.data?.pagination;

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
      <header className="space-y-3">
        <h1 className="font-heading text-3xl font-medium tracking-tight text-foreground">
          Shop
        </h1>
        <p className="max-w-2xl text-sm leading-relaxed text-muted">
          Browse our daily selection. Filter by category or search by name.
        </p>
      </header>

      <div className="mt-10 space-y-8">
        <section aria-label="Filters" className="space-y-6">
          {categoriesQuery.isPending ? (
            <p className="text-sm text-muted">Loading categories…</p>
          ) : categoriesQuery.isError ? (
            <p className="text-sm text-error">Failed to load categories.</p>
          ) : categories.length > 0 ? (
            <div className="flex flex-wrap gap-2">
              <button
                className={`rounded-md border px-4 py-2 text-sm font-medium transition-colors duration-default ease-default ${
                  params.category === undefined
                    ? "border-primary bg-primary-subtle text-foreground"
                    : "border-border bg-surface text-muted hover:border-border-strong hover:text-foreground"
                }`}
                onClick={() => setCategory(undefined)}
                type="button"
              >
                All
              </button>
              {categories.map((category) => (
                <button
                  key={category.id}
                  className={`rounded-md border px-4 py-2 text-sm font-medium transition-colors duration-default ease-default ${
                    params.category === category.id
                      ? "border-primary bg-primary-subtle text-foreground"
                      : "border-border bg-surface text-muted hover:border-border-strong hover:text-foreground"
                  }`}
                  onClick={() => setCategory(category.id)}
                  type="button"
                >
                  {category.name}
                </button>
              ))}
            </div>
          ) : null}

          <CatalogSearchForm
            key={params.search}
            defaultValue={params.search}
            onSearch={setSearch}
            trailing={
              hasActiveFilters ? (
                <Button onClick={clearFilters} type="button" variant="ghost">
                  Clear filters
                </Button>
              ) : null
            }
          />
        </section>

        <section aria-label="Products" className="space-y-6">
          {productsQuery.isPending ? (
            <p className="text-sm text-muted">Loading products…</p>
          ) : productsQuery.isError ? (
            <p className="text-sm text-error">Failed to load products.</p>
          ) : products.length === 0 ? (
            <p className="text-sm text-muted">
              {hasActiveFilters
                ? "No products match your filters."
                : "No products available."}
            </p>
          ) : (
            <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {products.map((product) => (
                <ProductCard key={product.id} product={product} />
              ))}
            </div>
          )}

          {pagination ? (
            <CatalogPagination
              onPageChange={setPage}
              page={params.page}
              totalPages={pagination.total_pages}
            />
          ) : null}
        </section>
      </div>
    </div>
  );
}
