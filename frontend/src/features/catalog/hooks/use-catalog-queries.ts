"use client";

import { useQuery } from "@tanstack/react-query";
import { z } from "zod";

import type {
  CatalogCategoriesListFilterInput,
  CatalogCombosListFilterInput,
  CatalogProductsListFilterInput,
} from "@/lib/schemas/catalog";

import {
  catalogCategoriesQueryOptions,
  catalogCombosQueryOptions,
  catalogProductQueryOptions,
  catalogProductsQueryOptions,
} from "./query-options";

export function useCatalogCategories(
  input?: CatalogCategoriesListFilterInput,
) {
  return useQuery(catalogCategoriesQueryOptions(input));
}

export function useCatalogProducts(input: CatalogProductsListFilterInput) {
  return useQuery(catalogProductsQueryOptions(input));
}

export function useCatalogProduct(
  id: string,
  options?: { enabled?: boolean },
) {
  const isValidId = z.string().uuid().safeParse(id).success;
  return useQuery({
    ...catalogProductQueryOptions(id, isValidId && (options?.enabled ?? true)),
    staleTime: 60_000,
  });
}

export function useCatalogCombos(input?: CatalogCombosListFilterInput) {
  return useQuery(catalogCombosQueryOptions(input));
}
