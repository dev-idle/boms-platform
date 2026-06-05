import { z } from "zod";

import { catalogSlugSchema } from "@/lib/validation/catalog";

const optionalUrl = z
  .string()
  .trim()
  .max(2048)
  .url("Enter a valid URL")
  .optional()
  .nullable();

export const managerCategorySchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1),
  slug: catalogSlugSchema,
  sort_order: z.number().int().min(0),
  is_active: z.boolean(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const categoryFormSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(255),
  slug: catalogSlugSchema,
  sort_order: z.coerce.number().int().min(0),
  is_active: z.boolean(),
});

export const categoryListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
  search: z.string().optional().default(""),
});

export const managerProductSchema = z.object({
  id: z.string().uuid(),
  category_id: z.string().uuid(),
  category_name: z.string().optional(),
  name: z.string().min(1),
  slug: catalogSlugSchema,
  description: z.string().nullable().optional(),
  price_cents: z.number().int().min(0),
  is_available: z.boolean(),
  image_url: z.string().nullable().optional(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const productFormSchema = z.object({
  category_id: z.string().uuid("Select a category"),
  name: z.string().trim().min(1, "Name is required").max(255),
  slug: catalogSlugSchema,
  description: z
    .string()
    .trim()
    .max(2000)
    .optional()
    .nullable(),
  price_cents: z.coerce.number().int().min(0, "Price must be zero or greater"),
  is_available: z.boolean(),
  image_url: optionalUrl,
});

export const productListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
  search: z.string().optional().default(""),
  category_id: z.string().optional().default(""),
});

export type ManagerCategory = z.infer<typeof managerCategorySchema>;
export type CategoryFormInput = z.infer<typeof categoryFormSchema>;
export type CategoryListFilterInput = z.infer<typeof categoryListFilterSchema>;

export type ManagerProduct = z.infer<typeof managerProductSchema>;
export type ProductFormInput = z.infer<typeof productFormSchema>;
export type ProductListFilterInput = z.infer<typeof productListFilterSchema>;

export type CategoriesListResult = {
  categories: ManagerCategory[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
};

export type ProductsListResult = {
  products: ManagerProduct[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
};
