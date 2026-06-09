"use client";

import { useQuery } from "@tanstack/react-query";
import { z } from "zod";

import {
  catalogCategoriesQueryOptions,
  catalogCombosQueryOptions,
  catalogProductQueryOptions,
  catalogProductsQueryOptions,
} from "./query-options";
import type {
  CatalogCategoriesListFilterInput,
  CatalogCombosListFilterInput,
  CatalogProductsListFilterInput,
} from "../schemas";

export { catalogQueryKeys } from "./query-options";
export {
  catalogCategoriesQueryOptions,
  catalogCombosQueryOptions,
  catalogProductQueryOptions,
  catalogProductsQueryOptions,
} from "./query-options";
export { useCatalogBrowseFilters } from "./use-catalog-browse-filters";

export function useCatalogCategories(
  input?: CatalogCategoriesListFilterInput,
) {
  return useQuery(catalogCategoriesQueryOptions(input));
}

export function useCatalogProducts(input: CatalogProductsListFilterInput) {
  return useQuery(catalogProductsQueryOptions(input));
}

export function useCatalogProduct(id: string) {
  const isValidId = z.string().uuid().safeParse(id).success;
  return useQuery(catalogProductQueryOptions(id, isValidId));
}

export function useCatalogCombos(input?: CatalogCombosListFilterInput) {
  return useQuery(catalogCombosQueryOptions(input));
}
