"use client";

import Link from "next/link";
import type { RefObject } from "react";

import { ROUTE } from "@/constants/routes";
import {
  buildCatalogBrowseHref,
  CATALOG_COMBOS_HEADING_ID,
  CatalogTile,
  CUSTOM_CAKE_BROWSE_HREF,
  useCombosOnSale,
} from "@/features/catalog";
import type { CatalogCategory } from "@/lib/schemas/catalog";
import { cn, plural } from "@/lib/utils";

type StorefrontHeaderCategoriesProps = {
  categories: CatalogCategory[];
  /** The first link, focused as the menu opens so Tab continues inside it. */
  firstLinkRef: RefObject<HTMLAnchorElement | null>;
  onNavigate: () => void;
  open: boolean;
  panelId: string;
  panelRef: RefObject<HTMLDivElement | null>;
};

/**
 * Every category with its photo and how many products it holds, opened from
 * the header. The toolbar keeps its few fixed links, so the menu holds any
 * number of categories without crowding the row. Below 64rem it is the whole
 * menu, so it also carries the links the toolbar shows from there.
 */
export function StorefrontHeaderCategories({
  categories,
  firstLinkRef,
  onNavigate,
  open,
  panelId,
  panelRef,
}: StorefrontHeaderCategoriesProps) {
  const productCount = categories.reduce((total, category) => total + category.product_count, 0);
  const combosOnSale = useCombosOnSale();

  return (
    <div
      ref={panelRef}
      className={cn(
        "storefront-header-menu",
        open && "storefront-header-menu--open",
      )}
      id={panelId}
      inert={!open || undefined}
    >
      <div className="storefront-container storefront-header-menu__inner">
        <div className="storefront-header-menu__head">
          <p className="storefront-header-menu__label">Categories</p>
          <Link className="storefront-header-menu__all" href={ROUTE.products} onClick={onNavigate} ref={firstLinkRef}>
            {productCount > 0 ? `All products · ${productCount}` : "All products"}
          </Link>
        </div>
        <ul className="storefront-header-menu__list">
          {categories.map((category) => (
            <li key={category.id}>
              <Link
                className="storefront-header-menu__link"
                href={buildCatalogBrowseHref({ category: category.id, page: 1 })}
                onClick={onNavigate}
              >
                <CatalogTile name={category.name} scale="row" url={category.image_url} />
                <span className="storefront-header-menu__copy">
                  <span className="storefront-header-menu__name">{category.name}</span>
                  <span className="storefront-header-menu__count">{plural(category.product_count, "product")}</span>
                </span>
              </Link>
            </li>
          ))}
        </ul>
        <ul className="storefront-header-menu__more">
          <li>
            <Link className="storefront-header-menu__row" href={CUSTOM_CAKE_BROWSE_HREF} onClick={onNavigate}>
              Custom cakes
            </Link>
          </li>
          {combosOnSale ? (
            <li>
              <Link
                className="storefront-header-menu__row"
                href={`${ROUTE.products}#${CATALOG_COMBOS_HEADING_ID}`}
                onClick={onNavigate}
              >
                Combos
              </Link>
            </li>
          ) : null}
        </ul>
      </div>
    </div>
  );
}
