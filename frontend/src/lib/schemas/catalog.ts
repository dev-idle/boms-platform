import { z } from "zod";

import { catalogSlugSchema } from "@/lib/validation/catalog";
import { productImageUrlsSchema } from "@/lib/validation/cloudinary";

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
  image_urls: productImageUrlsSchema.default([]),
});

export const catalogCategoriesListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(100),
});

export const catalogProductsListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(24),
  category_id: z.string().uuid().optional(),
  search: z.string().max(100).optional().default(""),
});

const comboItemSchema = z.object({
  product_id: z.string().uuid(),
  product_name: z.string().min(1),
  product_slug: catalogSlugSchema,
  quantity: z.number().int().min(1),
  price_cents: z.number().int().min(0),
});

export const catalogComboSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1),
  slug: catalogSlugSchema,
  price_cents: z.number().int().min(0),
  starts_at: z.string().datetime(),
  ends_at: z.string().datetime(),
  items: z.array(comboItemSchema),
});

export const catalogCombosListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(12),
});

export type CatalogCategory = z.infer<typeof catalogCategorySchema>;
export type CatalogProduct = z.infer<typeof catalogProductSchema>;
export type CatalogCategoriesListFilterInput = z.infer<
  typeof catalogCategoriesListFilterSchema
>;
export type CatalogProductsListFilterInput = z.infer<
  typeof catalogProductsListFilterSchema
>;
export type CatalogCombo = z.infer<typeof catalogComboSchema>;
export type CatalogCombosListFilterInput = z.infer<
  typeof catalogCombosListFilterSchema
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

export type CatalogCombosListResult = {
  combos: CatalogCombo[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
};
