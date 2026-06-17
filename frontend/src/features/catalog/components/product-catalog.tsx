"use client";

import { useMemo } from "react";

import { Button } from "@/components/ui/button";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { StorefrontAsyncPanel } from "@/components/ui/storefront-async-panel";

import { catalogProductFallbackImageUrl } from "@/constants/storefront-imagery";

import { toCatalogProductsFilter } from "../lib/catalog-browse-params";
import {
  useCatalogBrowseFilters,
  useCatalogCategories,
  useCatalogCombos,
  useCatalogProducts,
} from "../hooks";
import { CatalogCategoryFilters } from "./catalog-category-filters";
import { CatalogMenuPanel } from "./catalog-menu-panel";
import { CatalogPagination } from "./catalog-pagination";
import { ProductCard } from "./product-card";

const PRODUCTS_PAGE_SIZE = 24;

export function ProductCatalog() {
  const {
    params,
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
  const combosProbeQuery = useCatalogCombos({ page: 1, page_size: 1 });
  const productsQuery = useCatalogProducts(productsFilter);

  const categories = categoriesQuery.data?.categories ?? [];
  const products = productsQuery.data?.products ?? [];
  const pagination = productsQuery.data?.pagination;
  const showResultsMeta =
    !productsQuery.isPending && !productsQuery.isError && pagination;
  const activeSearch = params.search.trim();
  const showCombosLink =
    !combosProbeQuery.isPending &&
    (combosProbeQuery.data?.pagination.total ?? 0) > 0;
  const productsInitialLoading =
    productsQuery.isPending && productsQuery.data === undefined;
  const productsRefetching =
    productsQuery.isFetching && productsQuery.data !== undefined;
  const catalogInitialLoading =
    (categoriesQuery.isPending && categoriesQuery.data === undefined) ||
    productsInitialLoading;

  if (catalogInitialLoading) {
    return (
      <div className="catalog-page">
        <h1 className="sr-only">Shop</h1>
        <div className="storefront-container catalog-page__loading-shell">
          <InlineLoadingState className="catalog-page__loading" />
        </div>
      </div>
    );
  }

  return (
    <div className="catalog-page">
      <h1 className="sr-only">Shop</h1>

      <div className="storefront-container catalog-page__layout">
        <CatalogMenuPanel>
          {categoriesQuery.isError ? (
            <p className="text-sm text-error">Failed to load categories.</p>
          ) : (
            <CatalogCategoryFilters
              categories={categories}
              onSelectCategory={setCategory}
              selectedCategoryId={params.category}
              showCombosLink={showCombosLink}
            />
          )}
        </CatalogMenuPanel>

        <div className="catalog-page__content">
          {showResultsMeta ? (
            <p className="catalog-page__mobile-count">
              {pagination.total}
            </p>
          ) : null}

          {activeSearch ? (
            <p className="catalog-page__search-note">
              Showing results for &ldquo;{activeSearch}&rdquo;
            </p>
          ) : null}

          <StorefrontAsyncPanel
            className="catalog-page__results-panel"
            initialLoading={productsInitialLoading}
            overlayOnInitialLoad={false}
            refetching={productsRefetching}
          >
            <section aria-label="Products" className="catalog-page__results">
              {productsQuery.isError ? (
                <p className="text-sm text-error">Failed to load products.</p>
              ) : products.length === 0 ? (
                <div className="catalog-empty-state">
                  <p className="text-empty-title">
                    {hasActiveFilters
                      ? "No products match your filters"
                      : "No products available"}
                  </p>
                  <p className="catalog-empty-state__hint text-caption">
                    {hasActiveFilters
                      ? "Try another category or clear filters."
                      : "Check back soon — we refresh the menu daily."}
                  </p>
                  {hasActiveFilters ? (
                    <Button
                      onClick={clearFilters}
                      type="button"
                      variant="outline"
                    >
                      Clear filters
                    </Button>
                  ) : null}
                </div>
              ) : (
                <div className="catalog-product-grid">
                  {products.map((product, index) => (
                    <ProductCard
                      key={product.id}
                      fallbackImageUrl={catalogProductFallbackImageUrl(index)}
                      product={product}
                      showCategory={params.category === undefined}
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
          </StorefrontAsyncPanel>
        </div>
      </div>
    </div>
  );
}
