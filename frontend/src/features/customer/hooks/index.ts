"use client";

import { keepPreviousData, useQuery } from "@tanstack/react-query";

import { listCatalogCategories, listCatalogProducts } from "../api";
import {
  catalogCategoriesListFilterSchema,
  catalogProductsListFilterSchema,
  type CatalogCategoriesListFilterInput,
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
