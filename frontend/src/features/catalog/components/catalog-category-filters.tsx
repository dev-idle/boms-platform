import { cn } from "@/lib/utils";

import { CatalogCombosMenuLink } from "./catalog-combos-menu-link";

import type { CatalogCategory } from "@/lib/schemas/catalog";

type CatalogCategoryFiltersProps = {
  categories: CatalogCategory[];
  selectedCategoryId?: string;
  onSelectCategory: (categoryId: string | undefined) => void;
  showCombosLink?: boolean;
};

/** Vertical category list for the products page sidebar. */
export function CatalogCategoryFilters({
  categories,
  selectedCategoryId,
  onSelectCategory,
  showCombosLink = false,
}: CatalogCategoryFiltersProps) {
  if (categories.length === 0 && !showCombosLink) {
    return null;
  }

  return (
    <nav aria-label="Category filters" className="catalog-sidebar__nav">
      <p className="catalog-sidebar__label">Category</p>
      <ul className="catalog-sidebar__list">
        {categories.length > 0 ? (
          <li>
            <button
              aria-current={
                selectedCategoryId === undefined ? "true" : undefined
              }
              className={cn(
                "catalog-sidebar__link",
                selectedCategoryId === undefined &&
                  "catalog-sidebar__link--active",
              )}
              onClick={() => onSelectCategory(undefined)}
              type="button"
            >
              All products
            </button>
          </li>
        ) : null}
        {categories.map((category) => (
          <li key={category.id}>
            <button
              aria-current={
                selectedCategoryId === category.id ? "true" : undefined
              }
              className={cn(
                "catalog-sidebar__link",
                selectedCategoryId === category.id &&
                  "catalog-sidebar__link--active",
              )}
              onClick={() => onSelectCategory(category.id)}
              type="button"
            >
              {category.name}
            </button>
          </li>
        ))}
        {showCombosLink ? (
          <li>
            <CatalogCombosMenuLink className="catalog-sidebar__link catalog-sidebar__link--anchor" />
          </li>
        ) : null}
      </ul>
    </nav>
  );
}
