import type { CategoryListFilterInput, ProductListFilterInput } from "../schemas";

export const managerQueryKeys = {
  all: ["manager"] as const,
  categoriesRoot: ["manager", "categories"] as const,
  categories: (filter: CategoryListFilterInput) =>
    [...managerQueryKeys.categoriesRoot, filter] as const,
  category: (id: string) => [...managerQueryKeys.all, "category", id] as const,
  productsRoot: ["manager", "products"] as const,
  products: (filter: ProductListFilterInput) =>
    [...managerQueryKeys.productsRoot, filter] as const,
  product: (id: string) => [...managerQueryKeys.all, "product", id] as const,
};
