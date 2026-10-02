import { infiniteQueryOptions, keepPreviousData, queryOptions } from "@tanstack/react-query";

import {
  catalogCategoriesListFilterSchema,
  catalogCombosListFilterSchema,
  catalogProductsListFilterSchema,
  type CatalogCategoriesListFilterInput,
  type CatalogCombosListFilterInput,
  type CatalogProductsListFilterInput,
} from "@/lib/schemas/catalog";

import { CATALOG_CATEGORIES_PAGE_SIZE } from "@/constants/catalog";

import {
  getCatalogProduct,
  getProductReviews,
  listCatalogCategories,
  listCatalogCombos,
  listCatalogProducts,
} from "../api";

const catalogQueryKeys = {
  categoriesRoot: ["catalog", "categories"] as const,
  categories: (filter: CatalogCategoriesListFilterInput) =>
    [...catalogQueryKeys.categoriesRoot, filter] as const,
  productsRoot: ["catalog", "products"] as const,
  products: (filter: CatalogProductsListFilterInput) =>
    [...catalogQueryKeys.productsRoot, filter] as const,
  product: (id: string) => ["catalog", "product", id] as const,
  productReviews: (id: string) => ["catalog", "product", id, "reviews"] as const,
  combosRoot: ["catalog", "combos"] as const,
  combos: (filter: CatalogCombosListFilterInput) =>
    [...catalogQueryKeys.combosRoot, filter] as const,
};

export function catalogCategoriesQueryOptions(
  input: CatalogCategoriesListFilterInput = {
    page: 1,
    page_size: CATALOG_CATEGORIES_PAGE_SIZE,
  },
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

/** A product's published reviews, a page at a time: "Show more" reads those before the last shown. */
export function catalogProductReviewsQueryOptions(productId: string) {
  return infiniteQueryOptions({
    queryKey: catalogQueryKeys.productReviews(productId),
    queryFn: ({ pageParam }) => getProductReviews(productId, pageParam),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (page) => (page.has_more ? page.reviews.at(-1)?.id : undefined),
    staleTime: 60_000,
  });
}

export function catalogCombosQueryOptions(input: CatalogCombosListFilterInput) {
  const filter = catalogCombosListFilterSchema.parse(input);
  return queryOptions({
    queryKey: catalogQueryKeys.combos(filter),
    queryFn: () => listCatalogCombos(filter),
    placeholderData: keepPreviousData,
    staleTime: 60_000,
  });
}
