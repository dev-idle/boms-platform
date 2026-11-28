"use client";

import { useMemo, type ReactNode } from "react";

import { StorefrontPageHeader } from "@/components/layouts/storefront-page-header";
import { StorefrontEmptyState, StorefrontErrorState } from "@/components/layouts/storefront-state";
import { AsyncPanel } from "@/components/ui/async-panel";
import { Button } from "@/components/ui/button";
import { CATALOG_PRODUCTS_PAGE_SIZE } from "@/constants/catalog";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import type { CatalogProduct } from "@/lib/schemas/catalog";
import { plural } from "@/lib/utils";

import {
  useCatalogBrowseFilters,
  useCatalogCategories,
  useCatalogProducts,
  useCombosOnSale,
} from "../hooks";
import { toCatalogProductsFilter } from "../lib/catalog-browse-params";
import { groupByCategory } from "../lib/group-by-category";
import { CatalogCategoryFilters } from "./catalog-category-filters";
import { CatalogPagination } from "./catalog-pagination";
import { ProductCard } from "./product-card";
import { ProductCatalogLoading } from "./product-catalog-loading";

type ProductCatalogProps = {
  /** The save toggles and the purchase on each product's card, from the slice that owns them. */
  renderProductSaves?: (product: CatalogProduct) => ReactNode;
  renderProductPurchase?: (product: CatalogProduct) => ReactNode;
  /** The combo bundles, under the products in the same column. */
  combos?: ReactNode;
};

export function ProductCatalog({ renderProductSaves, renderProductPurchase, combos }: ProductCatalogProps) {
  const { params, setCategory, setPage, setSearch, clearFilters, hasActiveFilters } = useCatalogBrowseFilters();

  const productsFilter = useMemo(() => toCatalogProductsFilter(params, CATALOG_PRODUCTS_PAGE_SIZE), [params]);

  const categoriesQuery = useCatalogCategories();
  const showCombosLink = useCombosOnSale();
  const productsQuery = useCatalogProducts(productsFilter);

  const categories = categoriesQuery.data?.categories ?? [];
  const products = productsQuery.data?.products ?? [];
  const pagination = productsQuery.data?.pagination;
  const activeSearch = params.search.trim();
  const category = categories.find((candidate) => candidate.id === params.category);
  const { initialLoading: productsInitialLoading, refetching: productsRefetching } = getQuerySurface(productsQuery);

  if (categoriesQuery.isPending || productsInitialLoading) {
    return <ProductCatalogLoading />;
  }

  return (
    <div className="storefront-container storefront-page">
      <div className="catalog-page__top">
        <StorefrontPageHeader title={category?.name ?? PAGE_TITLES.shop} />
        {categoriesQuery.isError ? null : (
          <CatalogCategoryFilters
            categories={categories}
            onSelectCategory={setCategory}
            selectedCategoryId={params.category}
            showCombosLink={showCombosLink}
          />
        )}
      </div>

      <div className="catalog-page__body">
        {categoriesQuery.isError ? (
          <StorefrontErrorState
            onRetry={() => void categoriesQuery.refetch()}
            title="We could not load the categories"
          />
        ) : null}

        <div className="catalog-page__content">
          <div className="catalog-page__toolbar">
            {pagination && !productsQuery.isError && activeSearch ? (
              <p className="catalog-page__count">
                {plural(pagination.total, "product")} for &ldquo;{activeSearch}&rdquo;
              </p>
            ) : null}
            {activeSearch ? (
              <button
                className="storefront-text-action storefront-text-action--warning"
                onClick={() => setSearch("")}
                type="button"
              >
                Clear search
              </button>
            ) : null}
          </div>

          <AsyncPanel
            className="catalog-page__results-panel"
            initialLoading={productsInitialLoading}
            overlayOnInitialLoad={false}
            refetching={productsRefetching}
          >
            <section aria-label="Products">
              {productsQuery.isError ? (
                <StorefrontErrorState
                  onRetry={() => void productsQuery.refetch()}
                  title="We could not load the products"
                />
              ) : products.length === 0 ? (
                <StorefrontEmptyState
                  action={
                    // A search is cleared from the toolbar above, so it gets no second button here.
                    hasActiveFilters && !activeSearch ? (
                      <Button onClick={clearFilters} size="sm" type="button" variant="outline">
                        Clear filters
                      </Button>
                    ) : undefined
                  }
                  message={
                    activeSearch
                      ? "Check the spelling or clear the search."
                      : hasActiveFilters
                        ? "Try another category, or clear the filters."
                        : "Please check back later."
                  }
                  title={
                    activeSearch
                      ? `No products match “${activeSearch}”`
                      : hasActiveFilters
                        ? "No products match"
                        : "No products yet"
                  }
                />
              ) : hasActiveFilters ? (
                <>
                  {/* The cards' h3 titles sit under a section heading, as in the grouped view. */}
                  <h2 className="sr-only">Results</h2>
                  <div className="catalog-product-grid">
                    {products.map((product) => (
                      <ProductCard
                        key={product.id}
                        product={product}
                        purchase={renderProductPurchase?.(product)}
                        saves={renderProductSaves?.(product)}
                        showCategory={params.category === undefined}
                      />
                    ))}
                  </div>
                </>
              ) : (
                groupByCategory(products).map((section) => (
                  <section aria-labelledby={`catalog-section-${section.id}`} className="catalog-section" key={section.id}>
                    <div className="catalog-section__head">
                      <h2 className="catalog-section__title" id={`catalog-section-${section.id}`}>
                        {section.name}
                      </h2>
                      <span className="catalog-section__count">{plural(section.products.length, "product")}</span>
                    </div>
                    <div className="catalog-product-grid">
                      {section.products.map((product) => (
                        <ProductCard
                          key={product.id}
                          product={product}
                          purchase={renderProductPurchase?.(product)}
                          saves={renderProductSaves?.(product)}
                          showCategory={false}
                        />
                      ))}
                    </div>
                  </section>
                ))
              )}

              {pagination ? (
                <CatalogPagination onPageChange={setPage} page={params.page} totalPages={pagination.total_pages} />
              ) : null}
            </section>
          </AsyncPanel>

          {combos}
        </div>
      </div>
    </div>
  );
}
