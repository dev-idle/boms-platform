import Link from "next/link";

import type { CatalogCategory } from "@/lib/schemas/catalog";
import { plural } from "@/lib/utils";

import { buildCatalogBrowseHref } from "../lib/catalog-browse-params";
import { CatalogTile } from "./catalog-tile";

type HomeCategoryGridProps = {
  categories: CatalogCategory[];
};

export function HomeCategoryGrid({ categories }: HomeCategoryGridProps) {
  if (categories.length === 0) {
    return null;
  }

  return (
    <ul className="storefront-category-grid mt-10">
      {categories.map((category) => (
        <li key={category.id}>
          <Link
            className="storefront-category-tile"
            href={buildCatalogBrowseHref({ category: category.id, page: 1 })}
          >
            <CatalogTile name={category.name} scale="card" url={category.image_url} />
            <span>
              <span className="storefront-category-tile__name">{category.name}</span>
              <span className="storefront-category-tile__count">{plural(category.product_count, "product")}</span>
            </span>
          </Link>
        </li>
      ))}
    </ul>
  );
}
