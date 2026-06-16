import { z } from "zod";

import { catalogSlugSchema } from "@/lib/validation/catalog";
import { productImageUrlsResponseSchema, productImageUrlsSchema } from "@/lib/validation/cloudinary";

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
  image_urls: productImageUrlsResponseSchema,
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
  image_urls: productImageUrlsSchema,
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

const comboItemResponseSchema = z.object({
  product_id: z.string().uuid(),
  product_name: z.string().min(1),
  product_slug: catalogSlugSchema,
  quantity: z.number().int().min(1),
  price_cents: z.number().int().min(0),
});

const comboItemFormSchema = z.object({
  product_id: z.string().uuid("Select a product"),
  quantity: z.coerce.number().int().min(1, "Quantity must be at least 1"),
});

export const managerComboSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1),
  slug: catalogSlugSchema,
  price_cents: z.number().int().min(0),
  starts_at: z.string().datetime(),
  ends_at: z.string().datetime(),
  is_active: z.boolean(),
  items: z.array(comboItemResponseSchema),
  created_at: z.string(),
  updated_at: z.string(),
});

export const comboFormSchema = z
  .object({
    name: z.string().trim().min(1, "Name is required").max(255),
    slug: catalogSlugSchema,
    price_cents: z.coerce.number().int().min(0, "Price must be zero or greater"),
    starts_at: z.string().datetime(),
    ends_at: z.string().datetime(),
    is_active: z.boolean(),
    items: z.array(comboItemFormSchema).min(1, "Add at least one product"),
  })
  .refine((data) => new Date(data.ends_at) > new Date(data.starts_at), {
    message: "End time must be after start time",
    path: ["ends_at"],
  });

export const comboListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
  search: z.string().optional().default(""),
});

export const DISCOUNT_TYPE = {
  percent: "percent",
  fixedCents: "fixed_cents",
} as const;

export const managerDiscountCodeSchema = z.object({
  id: z.string().uuid(),
  code: z.string().min(3),
  discount_type: z.enum([DISCOUNT_TYPE.percent, DISCOUNT_TYPE.fixedCents]),
  value: z.number().int().min(1),
  min_order_cents: z.number().int().min(0).nullable().optional(),
  max_uses: z.number().int().min(1).nullable().optional(),
  used_count: z.number().int().min(0),
  starts_at: z.string().datetime(),
  ends_at: z.string().datetime(),
  is_active: z.boolean(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const discountCodeFormSchema = z
  .object({
    code: z.string().trim().min(3, "Code is required").max(64),
    discount_type: z.enum([DISCOUNT_TYPE.percent, DISCOUNT_TYPE.fixedCents]),
    value: z
      .union([
        z.undefined(),
        z
          .number({ invalid_type_error: "Value is required" })
          .int()
          .min(1, "Value must be at least 1"),
      ])
      .refine((val) => val !== undefined, {
        message: "Value is required",
      })
      .transform((val) => val as number),
    min_order_cents: z.coerce
      .number()
      .int()
      .min(0)
      .optional()
      .nullable(),
    max_uses: z.coerce
      .number()
      .int()
      .min(1)
      .optional()
      .nullable(),
    starts_at: z.string().datetime(),
    ends_at: z.string().datetime(),
    is_active: z.boolean(),
  })
  .refine((data) => new Date(data.ends_at) > new Date(data.starts_at), {
    message: "End time must be after start time",
    path: ["ends_at"],
  })
  .refine(
    (data) =>
      data.discount_type !== DISCOUNT_TYPE.percent ||
      (data.value >= 1 && data.value <= 100),
    {
      message: "Percent must be between 1 and 100",
      path: ["value"],
    },
  );

export const discountCodeListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
  search: z.string().optional().default(""),
});

export type ManagerCombo = z.infer<typeof managerComboSchema>;
export type ComboFormInput = z.infer<typeof comboFormSchema>;
export type ComboListFilterInput = z.infer<typeof comboListFilterSchema>;

export type ManagerDiscountCode = z.infer<typeof managerDiscountCodeSchema>;
export type DiscountCodeFormInput = z.infer<typeof discountCodeFormSchema>;
/** RHF defaults — `value` unset on create until the manager enters it. */
export type DiscountCodeFormValues = Omit<DiscountCodeFormInput, "value"> & {
  value?: number;
};
export type DiscountCodeListFilterInput = z.infer<
  typeof discountCodeListFilterSchema
>;

export type CombosListResult = {
  combos: ManagerCombo[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
};

export type DiscountCodesListResult = {
  discount_codes: ManagerDiscountCode[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
};
