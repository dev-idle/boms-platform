import Link from "next/link";

import { buildCatalogBrowseHref } from "../lib/catalog-browse-params";

import type { CatalogCategory } from "@/lib/schemas/catalog";

type HomeCategoriesProps = {
  categories: CatalogCategory[];
};

export function HomeCategories({ categories }: HomeCategoriesProps) {
  if (categories.length === 0) {
    return null;
  }

  return (
    <section
      aria-labelledby="home-categories-heading"
      className="storefront-section border-t border-border bg-bg"
    >
      <div className="storefront-container">
        <div className="text-center">
          <h2 className="text-h2" id="home-categories-heading">
            Shop <span className="italic text-matcha-500">by category</span>
          </h2>
          <p className="mt-2 text-sm text-muted">
            From morning viennoiserie to celebration cakes.
          </p>
        </div>

        <ul className="storefront-category-grid mt-8">
          {categories.map((category) => (
            <li key={category.id}>
              <Link
                className="storefront-category-tile"
                href={buildCatalogBrowseHref({ category: category.id, page: 1 })}
              >
                <span className="storefront-category-tile__name">
                  {category.name}
                </span>
                <span className="storefront-category-tile__cta">
                  Shop now
                  <span aria-hidden="true">→</span>
                </span>
              </Link>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
