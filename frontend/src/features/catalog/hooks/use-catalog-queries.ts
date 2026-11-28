"use client";

import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { z } from "zod";

import { CATALOG_COMBOS_PAGE_SIZE } from "@/constants/catalog";
import type {
  CatalogCategoriesListFilterInput,
  CatalogCombosListFilterInput,
  CatalogProductsListFilterInput,
} from "@/lib/schemas/catalog";

import {
  catalogCategoriesQueryOptions,
  catalogCombosQueryOptions,
  catalogProductQueryOptions,
  catalogProductReviewsQueryOptions,
  catalogProductsQueryOptions,
} from "./query-options";

export function useCatalogCategories(
  input?: CatalogCategoriesListFilterInput,
) {
  return useQuery(catalogCategoriesQueryOptions(input));
}

export function useCatalogProducts(
  input: CatalogProductsListFilterInput,
  options?: { enabled?: boolean },
) {
  return useQuery({
    ...catalogProductsQueryOptions(input),
    enabled: options?.enabled ?? true,
  });
}

export function useCatalogProduct(
  id: string,
  options?: { enabled?: boolean },
) {
  const isValidId = z.uuid().safeParse(id).success;
  return useQuery({
    ...catalogProductQueryOptions(id, isValidId && (options?.enabled ?? true)),
    staleTime: 60_000,
  });
}

export function useProductReviews(productId: string) {
  return useInfiniteQuery(catalogProductReviewsQueryOptions(productId));
}

export function useCatalogCombos(input: CatalogCombosListFilterInput) {
  return useQuery(catalogCombosQueryOptions(input));
}

/**
 * Whether any combo is on sale, read from the first page the combos section
 * shows, so the links to that section and the section share one request.
 */
export function useCombosOnSale(): boolean {
  const combosQuery = useCatalogCombos({ page: 1, page_size: CATALOG_COMBOS_PAGE_SIZE });
  return (combosQuery.data?.pagination.total ?? 0) > 0;
}
