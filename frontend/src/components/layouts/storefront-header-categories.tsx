"use client";

import Link from "next/link";
import type { RefObject } from "react";

import { ROUTE } from "@/constants/routes";
import { buildCatalogBrowseHref } from "@/features/catalog";
import type { CatalogCategory } from "@/lib/schemas/catalog";
import { cn } from "@/lib/utils";

type StorefrontHeaderCategoriesProps = {
  categories: CatalogCategory[];
  onNavigate: () => void;
  open: boolean;
  panelId: string;
  panelRef: RefObject<HTMLDivElement | null>;
};

/**
 * The full category list, opened from the header. The toolbar keeps three fixed links, so the
 * menu holds any number of categories without crowding the row.
 */
export function StorefrontHeaderCategories({
  categories,
  onNavigate,
  open,
  panelId,
  panelRef,
}: StorefrontHeaderCategoriesProps) {
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
        <p className="storefront-header-menu__label">Categories</p>
        <ul className="storefront-header-menu__list">
          {categories.map((category) => (
            <li key={category.id}>
              <Link
                className="storefront-header-menu__link"
                href={buildCatalogBrowseHref({ category: category.id, page: 1 })}
                onClick={onNavigate}
              >
                {category.name}
              </Link>
            </li>
          ))}
          <li>
            <Link
              className="storefront-header-menu__link storefront-header-menu__link--all"
              href={ROUTE.products}
              onClick={onNavigate}
            >
              Everything in the shop
            </Link>
          </li>
        </ul>
      </div>
    </div>
  );
}
