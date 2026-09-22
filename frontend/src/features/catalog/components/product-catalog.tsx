"use client";

import { useMemo } from "react";

import { Button } from "@/components/ui/button";
import { AsyncPanel } from "@/components/ui/async-panel";
import { StorefrontPageHeader } from "@/components/layouts/storefront-page-header";

import { catalogProductFallbackImageUrl } from "@/constants/storefront-imagery";

import { toCatalogProductsFilter } from "../lib/catalog-browse-params";
import {
  useCatalogBrowseFilters,
  useCatalogCategories,
  useCatalogCombos,
  useCatalogProducts,
} from "../hooks";
import { CatalogCategoryFilters } from "./catalog-category-filters";
import { CatalogPagination } from "./catalog-pagination";
import { ProductCard } from "./product-card";
import { ProductCatalogLoading } from "./product-catalog-loading";

const PRODUCTS_PAGE_SIZE = 24;

const CATALOG_PAGE_LEAD =
  "Everything we bake today, in one place. Availability updates through the morning as each batch leaves the oven.";

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
    return <ProductCatalogLoading />;
  }

  return (
    <div className="catalog-page">
      <div className="storefront-container catalog-page__intro">
        <StorefrontPageHeader
          eyebrow="The collection"
          lead={CATALOG_PAGE_LEAD}
          title="All products"
        />
      </div>

      <div className="storefront-container catalog-page__body">
        <aside aria-label="Filters" className="catalog-sidebar">
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
        </aside>

        <div className="catalog-page__content">
          <div className="catalog-page__toolbar">
            <p className="catalog-page__count">
              {showResultsMeta
                ? `${pagination.total} ${pagination.total === 1 ? "product" : "products"}`
                : null}
            </p>
          </div>

          {activeSearch ? (
            <p className="catalog-page__search-note">
              Showing results for &ldquo;{activeSearch}&rdquo;
            </p>
          ) : null}

          <AsyncPanel
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
                  <h2 className="text-empty-title">
                    {hasActiveFilters
                      ? "No products match your filters"
                      : "No products available"}
                  </h2>
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
          </AsyncPanel>
        </div>
      </div>
    </div>
  );
}
