import { z } from "zod";

import {
  browserRequestWithMeta,
} from "@/lib/browser-api-client";
import { parsePaginatedList } from "@/lib/pagination/parse-paginated-list";

import {
  catalogCategoriesListFilterSchema,
  catalogCategorySchema,
  catalogProductSchema,
  catalogProductsListFilterSchema,
  type CatalogCategoriesListFilterInput,
  type CatalogCategoriesListResult,
  type CatalogCategory,
  type CatalogProduct,
  type CatalogProductsListFilterInput,
  type CatalogProductsListResult,
} from "../schemas";

export async function listCatalogCategories(
  input: CatalogCategoriesListFilterInput = { page: 1, page_size: 100 },
): Promise<CatalogCategoriesListResult> {
  const filter = catalogCategoriesListFilterSchema.parse(input);
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  const result = await browserRequestWithMeta<CatalogCategory[]>(
    `/api/v1/catalog/categories?${params.toString()}`,
    { method: "GET", schema: z.array(catalogCategorySchema) },
  );
  const parsed = parsePaginatedList(result.data, result.meta, {
    page: filter.page,
    page_size: filter.page_size,
  });
  return {
    categories: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}

export async function listCatalogProducts(
  input: CatalogProductsListFilterInput,
): Promise<CatalogProductsListResult> {
  const filter = catalogProductsListFilterSchema.parse(input);
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  if (filter.category_id) {
    params.set("category_id", filter.category_id);
  }
  const result = await browserRequestWithMeta<CatalogProduct[]>(
    `/api/v1/catalog/products?${params.toString()}`,
    { method: "GET", schema: z.array(catalogProductSchema) },
  );
  const parsed = parsePaginatedList(result.data, result.meta, {
    page: filter.page,
    page_size: filter.page_size,
  });
  return {
    products: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}
