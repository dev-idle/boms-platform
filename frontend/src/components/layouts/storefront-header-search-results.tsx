"use client";

import Link from "next/link";
import type { RefObject } from "react";

import { ROUTE } from "@/constants/routes";
import { buildCatalogBrowseHref, catalogItemInitial } from "@/features/catalog";
import { primaryCatalogProductImageUrl } from "@/lib/cloudinary/config";
import type { CatalogProduct } from "@/lib/schemas/catalog";
import { cn } from "@/lib/utils";
import { formatPriceCents } from "@/lib/validation/catalog";

type StorefrontHeaderSearchResultsProps = {
  activeIndex: number;
  listboxId: string;
  loading: boolean;
  onNavigate: () => void;
  open: boolean;
  optionId: (index: number) => string;
  panelRef: RefObject<HTMLDivElement | null>;
  products: CatalogProduct[];
  query: string;
  total: number;
};

/** Live matches under the header: the catalogue answers while the shopper types. */
export function StorefrontHeaderSearchResults({
  activeIndex,
  listboxId,
  loading,
  onNavigate,
  open,
  optionId,
  panelRef,
  products,
  query,
  total,
}: StorefrontHeaderSearchResultsProps) {
  const seeAllIndex = products.length;
  const empty = !loading && products.length === 0;

  return (
    <div
      ref={panelRef}
      aria-busy={loading || undefined}
      className={cn(
        "storefront-header-results",
        open && "storefront-header-results--open",
      )}
      inert={!open || undefined}
    >
      <div className="storefront-container storefront-header-results__inner">
        <p className="storefront-header-results__head">
          {loading ? "Searching" : `Matches for “${query}”`}
        </p>

        {empty ? (
          <p className="storefront-header-results__empty">
            Nothing matches yet — try a shorter word, or browse the whole shop.
          </p>
        ) : null}

        <ul
          aria-label="Search suggestions"
          className="storefront-header-results__list"
          id={listboxId}
          role="listbox"
        >
          {products.map((product, index) => {
            const imageUrl = primaryCatalogProductImageUrl(
              product.image_urls,
              160,
            );
            return (
              <li key={product.id} role="presentation">
                <Link
                  aria-selected={index === activeIndex}
                  className={cn(
                    "storefront-header-results__item",
                    index === activeIndex &&
                      "storefront-header-results__item--active",
                  )}
                  href={ROUTE.productDetail(product.id)}
                  id={optionId(index)}
                  onClick={onNavigate}
                  role="option"
                  tabIndex={-1}
                >
                  <span aria-hidden className="storefront-header-results__thumb">
                    {imageUrl ? (
                      // eslint-disable-next-line @next/next/no-img-element -- catalog URLs are external manager-provided links
                      <img alt="" src={imageUrl} />
                    ) : (
                      <span className="storefront-header-results__initial">
                        {catalogItemInitial(product.name)}
                      </span>
                    )}
                  </span>
                  <span className="storefront-header-results__copy">
                    <span className="storefront-header-results__name">
                      {product.name}
                    </span>
                    <span className="storefront-header-results__category">
                      {product.category_name}
                    </span>
                  </span>
                  <span className="storefront-header-results__price text-price">
                    {formatPriceCents(product.price_cents)}
                  </span>
                </Link>
              </li>
            );
          })}

          <li role="presentation">
            <Link
              aria-selected={seeAllIndex === activeIndex}
              className={cn(
                "storefront-header-results__item storefront-header-results__all",
                seeAllIndex === activeIndex &&
                  "storefront-header-results__item--active",
              )}
              href={buildCatalogBrowseHref({ search: query, page: 1 })}
              id={optionId(seeAllIndex)}
              onClick={onNavigate}
              role="option"
              tabIndex={-1}
            >
              <span>
                {empty ? "Browse the whole shop" : "See every match"}
              </span>
              {empty ? null : (
                <span className="storefront-header-results__count text-price">
                  {total}
                </span>
              )}
            </Link>
          </li>
        </ul>

        {/* Desktop only: there is no key to press on a touch keyboard row. */}
        <p aria-hidden className="storefront-header-results__hint">
          <span>
            <kbd>↑</kbd>
            <kbd>↓</kbd> browse
          </span>
          <span>
            <kbd>↵</kbd> open
          </span>
          <span>
            <kbd>esc</kbd> close
          </span>
        </p>
      </div>
    </div>
  );
}
