import {
  dalListCatalogCategories,
  STOREFRONT_CATEGORY_PAGE_SIZE,
} from "@/lib/dal/catalog";

import { StorefrontHeader } from "./storefront-header";

/** Header with live category nav; degrades to "Shop all" + "Combos" if the catalog API is unreachable. */
export async function StorefrontCategoryHeader() {
  const categories = await dalListCatalogCategories(STOREFRONT_CATEGORY_PAGE_SIZE).catch(() => []);

  return <StorefrontHeader categories={categories} />;
}
