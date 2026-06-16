"use client";

import { useMemo } from "react";

import { Button } from "@/components/ui/button";
import {
  StorefrontCategoryPillsSkeleton,
  StorefrontProductGridSkeleton,
} from "@/components/ui/storefront-loading-skeletons";
import { cn } from "@/lib/utils";

import { catalogProductFallbackImageUrl } from "@/constants/storefront-imagery";

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

function filterPillClass(active: boolean) {
  return cn(
    "min-h-11 rounded-full border px-4 py-2 text-sm font-medium transition-[background-color,color,box-shadow] duration-standard ease-default",
    active
      ? "border-matcha-200 bg-matcha-100 text-matcha-700"
      : "border-border bg-surface text-muted hover:bg-mint hover:text-matcha-500",
  );
}

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
    <div className="storefront-container py-12">
      <header className="space-y-3">
        <h1>Shop</h1>
        <p className="text-body max-w-2xl">
          Browse our daily selection. Filter by category or search by name.
        </p>
      </header>

      <div className="mt-10 space-y-8">
        <section aria-label="Filters" className="space-y-6">
          {categoriesQuery.isPending ? (
            <StorefrontCategoryPillsSkeleton />
          ) : categoriesQuery.isError ? (
            <p className="text-sm text-error">Failed to load categories.</p>
          ) : categories.length > 0 ? (
            <div className="flex flex-wrap gap-2">
              <button
                className={filterPillClass(params.category === undefined)}
                onClick={() => setCategory(undefined)}
                type="button"
              >
                All
              </button>
              {categories.map((category) => (
                <button
                  key={category.id}
                  className={filterPillClass(params.category === category.id)}
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
            <StorefrontProductGridSkeleton />
          ) : productsQuery.isError ? (
            <p className="text-sm text-error">Failed to load products.</p>
          ) : products.length === 0 ? (
            <p className="text-sm text-muted">
              {hasActiveFilters
                ? "No products match your filters."
                : "No products available."}
            </p>
          ) : (
            <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {products.map((product, index) => (
                <ProductCard
                  key={product.id}
                  fallbackImageUrl={catalogProductFallbackImageUrl(index)}
                  product={product}
                />
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
