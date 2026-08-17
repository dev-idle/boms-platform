import Link from "next/link";

import { ROUTE } from "@/constants/routes";
import { buildCatalogBrowseHref } from "@/features/catalog/lib/catalog-browse-params";
import { CATALOG_COMBOS_HEADING_ID } from "@/features/catalog/lib/scroll-to-storefront-anchor";

import type { CatalogCategory } from "@/lib/schemas/catalog";

type StorefrontHeaderNavProps = {
  categories: CatalogCategory[];
};

/** Category strip under the header toolbar — horizontally scrollable on mobile. */
export function StorefrontHeaderNav({ categories }: StorefrontHeaderNavProps) {
  return (
    <nav aria-label="Shop categories" className="storefront-header-nav">
      <div className="storefront-container">
        <ul className="storefront-header-nav__list">
          <li>
            <Link className="storefront-header-nav__link" href={ROUTE.products}>
              Shop all
            </Link>
          </li>
          {categories.map((category) => (
            <li key={category.id}>
              <Link
                className="storefront-header-nav__link"
                href={buildCatalogBrowseHref({ category: category.id, page: 1 })}
              >
                {category.name}
              </Link>
            </li>
          ))}
          <li>
            <Link
              className="storefront-header-nav__link"
              href={`${ROUTE.products}#${CATALOG_COMBOS_HEADING_ID}`}
            >
              Combos
            </Link>
          </li>
        </ul>
      </div>
    </nav>
  );
}
