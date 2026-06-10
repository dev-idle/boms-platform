import { keepPreviousData, queryOptions } from "@tanstack/react-query";

import {
  catalogCategoriesListFilterSchema,
  catalogCombosListFilterSchema,
  catalogProductsListFilterSchema,
  type CatalogCategoriesListFilterInput,
  type CatalogCombosListFilterInput,
  type CatalogProductsListFilterInput,
} from "@/lib/schemas/catalog";

import {
  getCatalogProduct,
  listCatalogCategories,
  listCatalogCombos,
  listCatalogProducts,
} from "../api";

export const catalogQueryKeys = {
  categoriesRoot: ["catalog", "categories"] as const,
  categories: (filter: CatalogCategoriesListFilterInput) =>
    [...catalogQueryKeys.categoriesRoot, filter] as const,
  productsRoot: ["catalog", "products"] as const,
  products: (filter: CatalogProductsListFilterInput) =>
    [...catalogQueryKeys.productsRoot, filter] as const,
  product: (id: string) => ["catalog", "product", id] as const,
  combosRoot: ["catalog", "combos"] as const,
  combos: (filter: CatalogCombosListFilterInput) =>
    [...catalogQueryKeys.combosRoot, filter] as const,
};

const defaultCategoriesFilter: CatalogCategoriesListFilterInput = {
  page: 1,
  page_size: 100,
};

export function catalogCategoriesQueryOptions(
  input: CatalogCategoriesListFilterInput = defaultCategoriesFilter,
) {
  const filter = catalogCategoriesListFilterSchema.parse(input);
  return queryOptions({
    queryKey: catalogQueryKeys.categories(filter),
    queryFn: () => listCatalogCategories(filter),
  });
}

export function catalogProductsQueryOptions(
  input: CatalogProductsListFilterInput,
) {
  const filter = catalogProductsListFilterSchema.parse(input);
  return queryOptions({
    queryKey: catalogQueryKeys.products(filter),
    queryFn: () => listCatalogProducts(filter),
    placeholderData: keepPreviousData,
  });
}

export function catalogProductQueryOptions(id: string, enabled: boolean) {
  return queryOptions({
    queryKey: catalogQueryKeys.product(id),
    queryFn: () => getCatalogProduct(id),
    enabled,
    retry: false,
    staleTime: 60_000,
  });
}

const defaultCombosFilter: CatalogCombosListFilterInput = {
  page: 1,
  page_size: 12,
};

export function catalogCombosQueryOptions(
  input: CatalogCombosListFilterInput = defaultCombosFilter,
) {
  const filter = catalogCombosListFilterSchema.parse(input);
  return queryOptions({
    queryKey: catalogQueryKeys.combos(filter),
    queryFn: () => listCatalogCombos(filter),
    placeholderData: keepPreviousData,
  });
}
