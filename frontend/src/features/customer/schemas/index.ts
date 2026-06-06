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
  search: z.string().max(100).optional().default(""),
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

export type CatalogCombo = z.infer<typeof catalogComboSchema>;
export type CatalogCombosListFilterInput = z.infer<
  typeof catalogCombosListFilterSchema
>;

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

export const cartItemSchema = z.object({
  id: z.string().uuid(),
  line_type: z.enum(["product", "combo"]),
  product_id: z.string().uuid().nullable().optional(),
  combo_id: z.string().uuid().nullable().optional(),
  name: z.string(),
  slug: catalogSlugSchema,
  quantity: z.number().int().min(1).max(99),
  unit_price_cents: z.number().int().min(0),
  line_total_cents: z.number().int().min(0),
  is_available: z.boolean(),
});

export const cartDiscountSchema = z.object({
  code: z.string().min(1),
  discount_type: z.enum(["percent", "fixed_cents"]),
  value: z.number().int().min(1),
  discount_cents: z.number().int().min(0),
});

export const cartSchema = z.object({
  id: z.string().uuid(),
  items: z.array(cartItemSchema),
  subtotal_cents: z.number().int().min(0),
  discount: cartDiscountSchema.nullable().optional(),
  discount_cents: z.number().int().min(0),
  total_cents: z.number().int().min(0),
  checkout_ready: z.boolean(),
});

export const addCartItemInputSchema = z
  .object({
    product_id: z.string().uuid().optional(),
    combo_id: z.string().uuid().optional(),
    quantity: z.number().int().min(1).max(99).default(1),
  })
  .refine(
    (value) =>
      (value.product_id !== undefined) !== (value.combo_id !== undefined),
    { message: "Provide exactly one of product_id or combo_id" },
  );

export const updateCartItemInputSchema = z.object({
  quantity: z.number().int().min(1).max(99),
});

export const applyCartDiscountInputSchema = z.object({
  code: z.string().min(3).max(64),
});

export const orderItemSchema = z.object({
  id: z.string().uuid(),
  line_type: z.enum(["product", "combo"]),
  product_id: z.string().uuid().nullable().optional(),
  combo_id: z.string().uuid().nullable().optional(),
  name: z.string(),
  slug: catalogSlugSchema,
  quantity: z.number().int().min(1),
  unit_price_cents: z.number().int().min(0),
  line_total_cents: z.number().int().min(0),
});

export const orderSchema = z.object({
  id: z.string().uuid(),
  status: z.enum(["pending", "confirmed", "cancelled", "fulfilled"]),
  subtotal_cents: z.number().int().min(0),
  discount_cents: z.number().int().min(0),
  total_cents: z.number().int().min(0),
  discount_code_snapshot: z.string().nullable().optional(),
  items: z.array(orderItemSchema),
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
});

export const orderSummarySchema = z.object({
  id: z.string().uuid(),
  status: z.enum(["pending", "confirmed", "cancelled", "fulfilled"]),
  total_cents: z.number().int().min(0),
  item_count: z.number().int().min(0),
  created_at: z.string().datetime(),
});

export const ordersListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
});

export type Cart = z.infer<typeof cartSchema>;
export type CartItem = z.infer<typeof cartItemSchema>;
export type AddCartItemInput = z.infer<typeof addCartItemInputSchema>;
export type UpdateCartItemInput = z.infer<typeof updateCartItemInputSchema>;
export type ApplyCartDiscountInput = z.infer<typeof applyCartDiscountInputSchema>;
export type Order = z.infer<typeof orderSchema>;
export type OrderSummary = z.infer<typeof orderSummarySchema>;
export type OrdersListFilterInput = z.infer<typeof ordersListFilterSchema>;

export type OrdersListResult = {
  orders: OrderSummary[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
};
