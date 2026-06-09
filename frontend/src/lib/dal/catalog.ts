import "server-only";

import { z } from "zod";

import { getBomsApiClient } from "@/lib/api-client";
import {
  catalogCategorySchema,
  catalogProductSchema,
  type CatalogCategory,
  type CatalogProduct,
} from "@/lib/schemas/catalog";

export async function dalListCatalogCategories(
  pageSize = 12,
): Promise<CatalogCategory[]> {
  const client = getBomsApiClient();
  return client.request<CatalogCategory[]>(
    `/api/v1/catalog/categories?page=1&page_size=${pageSize}`,
    {
      method: "GET",
      schema: z.array(catalogCategorySchema),
      skipCookieForwarding: true,
    },
  );
}

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
