import type { CatalogProduct } from "@/lib/schemas/catalog";

export type CategorySection = {
  id: string;
  name: string;
  products: CatalogProduct[];
};

/** Products in one section per category, in the order the catalog lists them. */
export function groupByCategory(products: readonly CatalogProduct[]): CategorySection[] {
  const sections = new Map<string, CategorySection>();
  for (const product of products) {
    const section = sections.get(product.category_id);
    if (section) {
      section.products.push(product);
    } else {
      sections.set(product.category_id, {
        id: product.category_id,
        name: product.category_name,
        products: [product],
      });
    }
  }
  return [...sections.values()];
}
