import { z } from "zod";

import { ROUTE } from "@/constants/routes";

import type { CatalogProductsListFilterInput } from "@/lib/schemas/catalog";

const catalogBrowseParamsSchema = z.object({
  search: z
    .string()
    .max(100)
    .optional()
    .transform((value) => value?.trim() ?? ""),
  category: z
    .string()
    .uuid()
    .optional()
    .or(z.literal("").transform(() => undefined)),
  page: z.coerce.number().int().min(1).catch(1),
});

export type CatalogBrowseParams = z.infer<typeof catalogBrowseParamsSchema>;

type SearchParamsInput =
  | URLSearchParams
  | Record<string, string | string[] | undefined>;

function firstValue(
  value: string | string[] | undefined,
): string | undefined {
  if (Array.isArray(value)) {
    return value[0];
  }
  return value;
}

export function parseCatalogBrowseParams(
  input: SearchParamsInput,
): CatalogBrowseParams {
  const raw =
    input instanceof URLSearchParams
      ? Object.fromEntries(input.entries())
      : input;

  return catalogBrowseParamsSchema.parse({
    search: firstValue(raw.search),
    category: firstValue(raw.category),
    page: firstValue(raw.page),
  });
}

type BuildCatalogBrowseHrefInput = {
  search?: string;
  category?: string;
  page?: number;
};

export function buildCatalogBrowseHref(
  params: BuildCatalogBrowseHrefInput = {},
): string {
  const normalized = catalogBrowseParamsSchema.parse(params);
  const searchParams = new URLSearchParams();
  if (normalized.search) {
    searchParams.set("search", normalized.search);
  }
  if (normalized.category) {
    searchParams.set("category", normalized.category);
  }
  if (normalized.page > 1) {
    searchParams.set("page", String(normalized.page));
  }
  const query = searchParams.toString();
  return query ? `${ROUTE.products}?${query}` : ROUTE.products;
}

export function toCatalogProductsFilter(
  params: CatalogBrowseParams,
  pageSize: number,
): CatalogProductsListFilterInput {
  return {
    page: params.page,
    page_size: pageSize,
    category_id: params.category,
    search: params.search,
  };
}
