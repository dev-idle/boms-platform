import { HomeCategoryGrid, HomeSectionError } from "@/features/catalog";
import { dalListCatalogCategories, STOREFRONT_CATEGORY_PAGE_SIZE } from "@/lib/dal/catalog";

export async function HomeCategoryGridSection() {
  const categories = await dalListCatalogCategories(STOREFRONT_CATEGORY_PAGE_SIZE).catch(() => null);

  if (categories === null) {
    return <HomeSectionError title="We could not load the categories" />;
  }

  return <HomeCategoryGrid categories={categories} />;
}
