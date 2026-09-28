import { z } from "zod";

import { CATALOG_INTEGER_MAX, catalogSlugSchema } from "@/lib/validation/catalog";
import { apiDateTimeSchema } from "@/lib/validation/datetime";
import {
  catalogImageUrlResponseSchema,
  catalogImageUrlSchema,
  productImageUrlsResponseSchema,
  productImageUrlsSchema,
} from "@/lib/validation/cloudinary";

import {
  COMBO_BUNDLE_MIN_MESSAGE,
  isValidComboBundle,
} from "../lib/combo-bundle-rules";

/**
 * An amount the manager must enter, unset until they do — so the field shows its
 * hint instead of a zero nobody typed, and submitting without one says so.
 */
function requiredAmount(message: string, min: number, minMessage: string) {
  // `.optional()` keeps the field unset until the manager types — the form's
  // default omits it — while the refine below is what makes it required. A
  // union would report `invalid_union` instead, which reads as "Invalid input".
  return z
    .number({ error: message })
    .int(minMessage)
    .min(min, minMessage)
    .optional()
    .refine((value) => value !== undefined, { error: message })
    .transform((value) => value as number);
}

const NAME_MAX_MESSAGE = "Name must be at most 255 characters";

const priceAmount = requiredAmount(
  "Price is required",
  0,
  "Price must be zero or greater",
);

/** Where a category's products are made — backend `category.Station`. */
export const stationSchema = z.enum(["kitchen", "counter"]);

/** How each station reads in the manager's forms and tables. */
export const STATION_LABEL: Record<z.infer<typeof stationSchema>, string> = {
  kitchen: "Kitchen",
  counter: "Counter",
};

/** A product's notice is bounded like backend `product.MaxLeadTime`: a week. */
export const MAX_PRODUCT_LEAD_MINUTES = 7 * 24 * 60;

export const managerCategorySchema = z.object({
  id: z.uuid(),
  name: z.string().min(1),
  slug: catalogSlugSchema,
  sort_order: z.number().int().min(0),
  is_active: z.boolean(),
  station: stationSchema,
  created_at: z.string(),
  updated_at: z.string(),
});

export const categoryFormSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(255, NAME_MAX_MESSAGE),
  slug: catalogSlugSchema,
  sort_order: z
    .number()
    .int()
    .min(0)
    .max(CATALOG_INTEGER_MAX, "Sort order is too large"),
  is_active: z.boolean(),
  station: stationSchema,
});

export const categoryListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
  search: z.string().optional().default(""),
});

export const managerProductSchema = z.object({
  id: z.uuid(),
  category_id: z.uuid(),
  category_name: z.string().optional(),
  name: z.string().min(1),
  slug: catalogSlugSchema,
  description: z.string().nullable().optional(),
  price_cents: z.number().int().min(0),
  is_active: z.boolean(),
  lead_time_minutes: z.number().int().min(0),
  image_urls: productImageUrlsResponseSchema,
  created_at: z.string(),
  updated_at: z.string(),
});

export const productFormSchema = z.object({
  category_id: z.uuid("Select a category"),
  name: z.string().trim().min(1, "Name is required").max(255, NAME_MAX_MESSAGE),
  slug: catalogSlugSchema,
  description: z
    .string()
    .trim()
    .max(2000, "Description must be at most 2000 characters")
    .optional()
    .nullable(),
  price_cents: priceAmount,
  is_active: z.boolean(),
  lead_time_minutes: z
    .number({ error: "Enter the notice in minutes" })
    .int()
    .min(0)
    .max(MAX_PRODUCT_LEAD_MINUTES, `At most ${MAX_PRODUCT_LEAD_MINUTES} minutes (7 days)`),
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
/** RHF defaults — `price_cents` unset on create until the manager enters it. */
export type ProductFormValues = z.input<typeof productFormSchema>;
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
  product_id: z.uuid(),
  product_name: z.string().min(1),
  product_slug: catalogSlugSchema,
  quantity: z.number().int().min(1),
  price_cents: z.number().int().min(0),
});

const comboItemFormSchema = z.object({
  product_id: z.uuid("Select a product"),
  quantity: z
    .number()
    .int()
    .min(1, "Quantity must be at least 1")
    .max(CATALOG_INTEGER_MAX, "Quantity is too large"),
});

export const managerComboSchema = z.object({
  id: z.uuid(),
  name: z.string().min(1),
  slug: catalogSlugSchema,
  price_cents: z.number().int().min(0),
  image_url: catalogImageUrlResponseSchema,
  starts_at: apiDateTimeSchema,
  ends_at: apiDateTimeSchema,
  is_active: z.boolean(),
  items: z.array(comboItemResponseSchema),
  created_at: z.string(),
  updated_at: z.string(),
});

export const comboFormSchema = z
  .object({
    name: z.string().trim().min(1, "Name is required").max(255, NAME_MAX_MESSAGE),
    slug: catalogSlugSchema,
    price_cents: priceAmount,
    image_url: catalogImageUrlSchema,
    starts_at: apiDateTimeSchema,
    ends_at: apiDateTimeSchema,
    is_active: z.boolean(),
    items: z.array(comboItemFormSchema).min(1, "Add at least one product"),
  })
  .refine((data) => isValidComboBundle(data.items), {
    error: COMBO_BUNDLE_MIN_MESSAGE,
    path: ["items"],
  })
  .refine((data) => new Date(data.ends_at) > new Date(data.starts_at), {
    error: "End time must be after start time",
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
  id: z.uuid(),
  code: z.string().min(3),
  discount_type: z.enum([DISCOUNT_TYPE.percent, DISCOUNT_TYPE.fixedCents]),
  value: z.number().int().min(1),
  min_order_cents: z.number().int().min(0).nullable().optional(),
  max_uses: z.number().int().min(1).nullable().optional(),
  max_discount_cents: z.number().int().min(1).nullable().optional(),
  used_count: z.number().int().min(0),
  starts_at: apiDateTimeSchema,
  ends_at: apiDateTimeSchema,
  is_active: z.boolean(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const discountCodeFormSchema = z
  .object({
    code: z
      .string()
      .trim()
      .min(3, "Code is required")
      .max(64, "Code must be at most 64 characters"),
    discount_type: z.enum([DISCOUNT_TYPE.percent, DISCOUNT_TYPE.fixedCents]),
    value: requiredAmount("Value is required", 1, "Value must be at least 1"),
    min_order_cents: z.number().int().min(0).optional().nullable(),
    max_uses: z
      .number()
      .int()
      .min(1)
      .max(CATALOG_INTEGER_MAX, "Maximum uses is too large")
      .optional()
      .nullable(),
    max_discount_cents: z.number().int().min(1).optional().nullable(),
    starts_at: apiDateTimeSchema,
    ends_at: apiDateTimeSchema,
    is_active: z.boolean(),
  })
  .refine((data) => new Date(data.ends_at) > new Date(data.starts_at), {
    error: "End time must be after start time",
    path: ["ends_at"],
  })
  .refine(
    (data) =>
      data.discount_type !== DISCOUNT_TYPE.percent ||
      (data.value >= 1 && data.value <= 100),
    {
      error: "Percent must be between 1 and 100",
      path: ["value"],
    },
  )
  .refine(
    (data) =>
      data.discount_type !== DISCOUNT_TYPE.fixedCents ||
      data.max_discount_cents == null,
    {
      error: "Max discount cap applies to percent discounts only",
      path: ["max_discount_cents"],
    },
  );

export const discountCodeListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
  search: z.string().optional().default(""),
});

export type ManagerCombo = z.infer<typeof managerComboSchema>;
export type ComboFormInput = z.infer<typeof comboFormSchema>;
/** RHF defaults — `price_cents` unset on create until the manager enters it. */
export type ComboFormValues = z.input<typeof comboFormSchema>;
export type ComboListFilterInput = z.infer<typeof comboListFilterSchema>;

export type ManagerDiscountCode = z.infer<typeof managerDiscountCodeSchema>;
export type DiscountCodeFormInput = z.infer<typeof discountCodeFormSchema>;
/** RHF defaults — `value` unset on create until the manager enters it. */
export type DiscountCodeFormValues = z.input<typeof discountCodeFormSchema>;
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
