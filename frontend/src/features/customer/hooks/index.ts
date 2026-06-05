"use client";

import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { z } from "zod";

import {
  getCatalogProduct,
  listCatalogCategories,
  listCatalogCombos,
  listCatalogProducts,
} from "../api";
import {
  catalogCategoriesListFilterSchema,
  catalogCombosListFilterSchema,
  catalogProductsListFilterSchema,
  type CatalogCategoriesListFilterInput,
  type CatalogCombosListFilterInput,
  type CatalogProductsListFilterInput,
} from "../schemas";
import { customerQueryKeys } from "./query-options";

export { customerQueryKeys } from "./query-options";

const defaultCategoriesFilter: CatalogCategoriesListFilterInput = {
  page: 1,
  page_size: 100,
};

export function useCatalogCategories(
  input: CatalogCategoriesListFilterInput = defaultCategoriesFilter,
) {
  const filter = catalogCategoriesListFilterSchema.parse(input);
  return useQuery({
    queryKey: customerQueryKeys.catalogCategories(filter),
    queryFn: () => listCatalogCategories(filter),
  });
}

export function useCatalogProducts(input: CatalogProductsListFilterInput) {
  const filter = catalogProductsListFilterSchema.parse(input);
  return useQuery({
    queryKey: customerQueryKeys.catalogProducts(filter),
    queryFn: () => listCatalogProducts(filter),
    placeholderData: keepPreviousData,
  });
}

export function useCatalogProduct(id: string) {
  const isValidId = z.string().uuid().safeParse(id).success;
  return useQuery({
    queryKey: customerQueryKeys.catalogProduct(id),
    queryFn: () => getCatalogProduct(id),
    enabled: isValidId,
    retry: false,
  });
}

const defaultCombosFilter: CatalogCombosListFilterInput = {
  page: 1,
  page_size: 12,
};

export function useCatalogCombos(
  input: CatalogCombosListFilterInput = defaultCombosFilter,
) {
  const filter = catalogCombosListFilterSchema.parse(input);
  return useQuery({
    queryKey: customerQueryKeys.catalogCombos(filter),
    queryFn: () => listCatalogCombos(filter),
    placeholderData: keepPreviousData,
  });
}

