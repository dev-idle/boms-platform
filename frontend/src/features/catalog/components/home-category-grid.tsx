import Link from "next/link";

import type { CatalogCategory } from "@/lib/schemas/catalog";

import { buildCatalogBrowseHref } from "../lib/catalog-browse-params";

type HomeCategoryGridProps = {
  categories: CatalogCategory[];
};

export function HomeCategoryGrid({ categories }: HomeCategoryGridProps) {
  if (categories.length === 0) {
    return null;
  }

  return (
    <ul className="storefront-category-grid mt-10">
      {categories.map((category, index) => (
        <li key={category.id}>
          <Link
            className="storefront-category-tile"
            href={buildCatalogBrowseHref({ category: category.id, page: 1 })}
          >
            <span className="storefront-category-tile__ordinal">
              {String(index + 1).padStart(2, "0")}
            </span>
            <span className="storefront-category-tile__body">
              <span className="storefront-category-tile__name">
                {category.name}
              </span>
              <span className="storefront-category-tile__cta">
                Explore →
              </span>
            </span>
          </Link>
        </li>
      ))}
    </ul>
  );
}
