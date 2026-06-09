import { z } from "zod";

import {
  browserRequest,
  browserRequestWithMeta,
} from "@/lib/browser-api-client";
import { parsePaginatedList } from "@/lib/pagination/parse-paginated-list";

import {
  catalogCategoriesListFilterSchema,
  catalogCategorySchema,
  catalogComboSchema,
  catalogCombosListFilterSchema,
  catalogProductSchema,
  catalogProductsListFilterSchema,
  type CatalogCategoriesListFilterInput,
  type CatalogCategoriesListResult,
  type CatalogCategory,
  type CatalogCombo,
  type CatalogCombosListFilterInput,
  type CatalogCombosListResult,
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
    {
      method: "GET",
      schema: z.array(catalogCategorySchema),
      skipRefreshRetry: true,
    },
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
  if (filter.search) {
    params.set("search", filter.search);
  }
  const result = await browserRequestWithMeta<CatalogProduct[]>(
    `/api/v1/catalog/products?${params.toString()}`,
    {
      method: "GET",
      schema: z.array(catalogProductSchema),
      skipRefreshRetry: true,
    },
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

export async function getCatalogProduct(id: string): Promise<CatalogProduct> {
  const parsedId = z.string().uuid().parse(id);
  return browserRequest<CatalogProduct>(`/api/v1/catalog/products/${parsedId}`, {
    method: "GET",
    schema: catalogProductSchema,
    skipRefreshRetry: true,
  });
}

export async function listCatalogCombos(
  input: CatalogCombosListFilterInput,
): Promise<CatalogCombosListResult> {
  const filter = catalogCombosListFilterSchema.parse(input);
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  const result = await browserRequestWithMeta<CatalogCombo[]>(
    `/api/v1/catalog/combos?${params.toString()}`,
    {
      method: "GET",
      schema: z.array(catalogComboSchema),
      skipRefreshRetry: true,
    },
  );
  const parsed = parsePaginatedList(result.data, result.meta, {
    page: filter.page,
    page_size: filter.page_size,
  });
  return {
    combos: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}
