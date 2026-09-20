import { HomeCategoryGrid } from "@/features/catalog";
import { dalListCatalogCategories, STOREFRONT_CATEGORY_PAGE_SIZE } from "@/lib/dal/catalog";

export async function HomeCategoryGridSection() {
  const categories = await dalListCatalogCategories(STOREFRONT_CATEGORY_PAGE_SIZE).catch(() => []);

  return <HomeCategoryGrid categories={categories} />;
}
