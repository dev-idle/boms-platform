import "server-only";

import { cache } from "react";
import { z } from "zod";

import { getBomsApiClient } from "@/lib/api-client";
import {
  catalogCategorySchema,
  catalogProductSchema,
  type CatalogCategory,
  type CatalogProduct,
} from "@/lib/schemas/catalog";

/**
 * Category count shared by the header nav and home tiles — both callers must
 * pass the same value or the per-request `cache()` dedup splits into two fetches.
 */
export const STOREFRONT_CATEGORY_PAGE_SIZE = 8;

/** Products in the home "Fresh from the oven" grid. */
export const HOME_FEATURED_PRODUCT_COUNT = 8;

/** Request-deduped: the header nav and the home page share one fetch. */
export const dalListCatalogCategories = cache(
  async (pageSize: number): Promise<CatalogCategory[]> => {
    const client = getBomsApiClient();
    return client.request<CatalogCategory[]>(
      `/api/v1/catalog/categories?page=1&page_size=${pageSize}`,
      {
        method: "GET",
        schema: z.array(catalogCategorySchema),
        skipCookieForwarding: true,
      },
    );
  },
);

export async function dalListCatalogProducts(
  page = 1,
  pageSize = 8,
): Promise<CatalogProduct[]> {
  const client = getBomsApiClient();
  return client.request<CatalogProduct[]>(
    `/api/v1/catalog/products?page=${page}&page_size=${pageSize}`,
    {
      method: "GET",
      schema: z.array(catalogProductSchema),
      skipCookieForwarding: true,
    },
  );
}

export async function dalGetCatalogProduct(id: string): Promise<CatalogProduct> {
  const parsedId = z.string().uuid().parse(id);
  const client = getBomsApiClient();
  return client.request<CatalogProduct>(
    `/api/v1/catalog/products/${parsedId}`,
    {
      method: "GET",
      schema: catalogProductSchema,
      skipCookieForwarding: true,
    },
  );
}
