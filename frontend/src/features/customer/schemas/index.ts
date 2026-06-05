import { z } from "zod";

import { catalogSlugSchema } from "@/lib/validation/catalog";

export const catalogCategorySchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1),
  slug: catalogSlugSchema,
  sort_order: z.number().int().min(0),
});

export const catalogProductSchema = z.object({
  id: z.string().uuid(),
  category_id: z.string().uuid(),
  category_name: z.string(),
  category_slug: catalogSlugSchema,
  name: z.string().min(1),
  slug: catalogSlugSchema,
  description: z.string().nullable().optional(),
  price_cents: z.number().int().min(0),
  image_url: z.string().nullable().optional(),
});

export const catalogCategoriesListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(100),
});

export const catalogProductsListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(24),
  category_id: z.string().uuid().optional(),
});

export type CatalogCategory = z.infer<typeof catalogCategorySchema>;
export type CatalogProduct = z.infer<typeof catalogProductSchema>;
export type CatalogCategoriesListFilterInput = z.infer<
  typeof catalogCategoriesListFilterSchema
>;
export type CatalogProductsListFilterInput = z.infer<
  typeof catalogProductsListFilterSchema
>;

export type CatalogCategoriesListResult = {
  categories: CatalogCategory[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
};

export type CatalogProductsListResult = {
  products: CatalogProduct[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
};
