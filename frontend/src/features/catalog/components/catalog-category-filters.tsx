import { cn } from "@/lib/utils";

import { CatalogCombosMenuLink } from "./catalog-combos-menu-link";

import type { CatalogCategory } from "@/lib/schemas/catalog";

type CatalogCategoryFiltersProps = {
  categories: CatalogCategory[];
  selectedCategoryId?: string;
  onSelectCategory: (categoryId: string | undefined) => void;
  showCombosLink?: boolean;
};

/** One row of category chips over the shop, each with how many products it holds. */
export function CatalogCategoryFilters({
  categories,
  selectedCategoryId,
  onSelectCategory,
  showCombosLink = false,
}: CatalogCategoryFiltersProps) {
  if (categories.length === 0 && !showCombosLink) {
    return null;
  }

  const total = categories.reduce((sum, category) => sum + category.product_count, 0);
  const chips = [
    { id: undefined, name: "All", count: total },
    ...categories.map((category) => ({ id: category.id, name: category.name, count: category.product_count })),
  ];

  return (
    <nav aria-label="Category filters" className="catalog-chips">
      <ul className="catalog-chips__list">
        {categories.length > 0
          ? chips.map((chip) => {
              const selected = selectedCategoryId === chip.id;
              return (
                <li key={chip.id ?? "all"}>
                  <button
                    aria-pressed={selected}
                    className={cn("catalog-chip", selected && "catalog-chip--active")}
                    onClick={() => onSelectCategory(chip.id)}
                    type="button"
                  >
                    {chip.name}
                    <span className="catalog-chip__count">{chip.count}</span>
                  </button>
                </li>
              );
            })
          : null}
        {showCombosLink ? (
          <li>
            <CatalogCombosMenuLink />
          </li>
        ) : null}
      </ul>
    </nav>
  );
}
