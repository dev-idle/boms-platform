import { cn } from "@/lib/utils";

import { CatalogCombosMenuLink } from "./catalog-combos-menu-link";

import type { CatalogCategory } from "@/lib/schemas/catalog";

type CatalogCategoryFiltersProps = {
  categories: CatalogCategory[];
  selectedCategoryId?: string;
  onSelectCategory: (categoryId: string | undefined) => void;
  showCombosLink?: boolean;
};

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
    <nav aria-label="Menu" className="catalog-menu__nav">
      <ul className="catalog-menu__list">
        {categories.length > 0 ? (
          <li>
            <button
              aria-current={
                selectedCategoryId === undefined ? "page" : undefined
              }
              className={cn(
                "catalog-menu__link",
                selectedCategoryId === undefined && "catalog-menu__link--active",
              )}
              onClick={() => onSelectCategory(undefined)}
              type="button"
            >
              All
            </button>
          </li>
        ) : null}
        {categories.map((category) => (
          <li key={category.id}>
            <button
              aria-current={
                selectedCategoryId === category.id ? "page" : undefined
              }
              className={cn(
                "catalog-menu__link",
                selectedCategoryId === category.id &&
                  "catalog-menu__link--active",
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
            <CatalogCombosMenuLink />
          </li>
        ) : null}
      </ul>
    </nav>
  );
}
