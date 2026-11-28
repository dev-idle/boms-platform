import { z } from "zod";

import { clockTimeSchema, clockToMinutes } from "@/lib/validation/clock";
import { reasonSchema } from "@/lib/validation/reason";

import { productOptionGroupSchema } from "@/lib/schemas/catalog";
import { incidentTypeSchema, orderIncidentSchema } from "@/lib/schemas/incident";
import { orderCodeSchema, orderStatusSchema } from "@/lib/schemas/order";
import { averageRatingSchema, MAX_RATING, ratingSchema, reviewStatusSchema } from "@/lib/schemas/review";
import { stationSchema } from "@/lib/schemas/ticket";
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

/** A product's notice is bounded like backend `product.MaxLeadTime`: a week. */
export const MAX_PRODUCT_LEAD_MINUTES = 7 * 24 * 60;

/** How many options one product offers, as backend `product.MaxOptions`. */
export const MAX_PRODUCT_OPTIONS = 30;

/** The most an option adds to the price, as backend `product.MaxPriceDeltaCents`. */
const MAX_OPTION_PRICE_DELTA_CENTS = 100_000;

const managerProductOptionSchema = z.object({
  id: z.uuid(),
  group: productOptionGroupSchema,
  label: z.string().min(1),
  price_delta_cents: z.number().int().min(0),
  is_active: z.boolean(),
});

/** One option a manager offers; one without an id is new, and one left out is retired. */
const productOptionFormSchema = z.object({
  id: z.uuid().optional(),
  group: productOptionGroupSchema,
  label: z
    .string()
    .trim()
    .min(1, "Name the option")
    .max(60, "At most 60 characters")
    .refine((value) => !/[\p{Cc}\p{Cf}]/u.test(value), "Use plain text only"),
  price_delta_cents: requiredAmount("Enter the added price", 0, "The added price is zero or more").refine(
    (value) => value <= MAX_OPTION_PRICE_DELTA_CENTS,
    "At most $1,000.00",
  ),
  is_active: z.boolean(),
});

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
  is_customizable: z.boolean(),
  /** On the product's detail only; retired options left out. */
  options: z.array(managerProductOptionSchema).default([]),
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
  is_customizable: z.boolean(),
  options: z.array(productOptionFormSchema).max(MAX_PRODUCT_OPTIONS, `At most ${MAX_PRODUCT_OPTIONS} options`),
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
  max_uses_per_customer: z.number().int().min(1).nullable().optional(),
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
    max_uses_per_customer: z
      .number()
      .int()
      .min(1)
      .max(CATALOG_INTEGER_MAX, "Uses per customer is too large")
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

/** GET /manager/reviews — a review as a manager reads it; the moderator is null on one no manager has hidden. */
export const managerReviewSchema = z.object({
  id: z.uuid(),
  order_code: orderCodeSchema,
  product_name: z.string().min(1),
  rating: ratingSchema,
  comment: z.string().min(1).nullable(),
  status: reviewStatusSchema,
  created_at: apiDateTimeSchema,
  customer_name: z.string().nullable(),
  moderator_name: z.string().nullable(),
});

export const reviewListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
  status: reviewStatusSchema.optional(),
});

/**
 * GET /manager/reviews/summary — every review not hidden added up: how many,
 * their average, how many gave each rating (`stars[0]` one star), and each
 * product's, the lowest average first.
 */
export const reviewSummarySchema = z.object({
  review_count: z.number().int().min(0),
  average_rating: averageRatingSchema.nullable(),
  stars: z.array(z.number().int().min(0)).length(MAX_RATING),
  products: z.array(
    z.object({
      product_id: z.uuid(),
      product_name: z.string().min(1),
      review_count: z.number().int().min(1),
      average_rating: averageRatingSchema,
    }),
  ),
});

/** GET /manager/incidents — an incident with its order and who recorded it, null when the system did. */
export const managerIncidentSchema = orderIncidentSchema.extend({
  order_code: orderCodeSchema,
  pickup_at: apiDateTimeSchema.nullable(),
  actor_name: z.string().nullable(),
});

/** A page of the incidents of the bakery week starting on the Monday `week`, narrowed to one type when one is given. */
export const incidentListFilterSchema = z.object({
  week: z.iso.date(),
  type: incidentTypeSchema.optional(),
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
});

/** GET /manager/incidents/summary — how many incidents of each type a week holds; a type with none is left out. */
export const incidentSummarySchema = z.object({
  types: z.array(z.object({ type: incidentTypeSchema, count: z.number().int().min(1) })),
});

const cents = z.number().int();
const count = z.number().int().min(0);

/** The bakery days a sales report covers, and the periods it adds up by. */
export const salesReportFilterSchema = z.object({
  from: z.iso.date(),
  to: z.iso.date(),
  group: z.enum(["day", "week", "month"]),
});

/**
 * GET /manager/reports/sales — the money taken and given back in the range,
 * in total and by period; what the orders kept (paid in it and not
 * cancelled) sold by item and category, and the discounts they used; and how
 * long orders took to make. Amounts in cents; a refund-heavy period nets
 * below zero.
 */
export const salesReportSchema = z.object({
  from: z.iso.date(),
  to: z.iso.date(),
  group: salesReportFilterSchema.shape.group,
  totals: z.object({
    orders: count,
    gross_cents: cents,
    refunds_cents: cents,
    net_cents: cents,
    average_order_cents: cents.nullable(),
  }),
  periods: z.array(
    z.object({
      start: z.iso.date(),
      orders: count,
      gross_cents: cents,
      refunds_cents: cents,
      net_cents: cents,
    }),
  ),
  items: z.array(
    z.object({
      kind: z.enum(["product", "combo"]),
      id: z.uuid(),
      name: z.string().min(1),
      quantity: count,
      sales_cents: cents,
    }),
  ),
  categories: z.array(
    z.object({
      category_id: z.uuid().nullable(),
      name: z.string().min(1).nullable(),
      quantity: count,
      sales_cents: cents,
    }),
  ),
  discounts: z.object({
    orders: count,
    discounted_orders: count,
    discount_cents: cents,
    codes: z.array(z.object({ code: z.string().min(1), orders: count, discount_cents: cents })),
  }),
  production: z.object({ orders: count, average_minutes: count.nullable() }),
});

/**
 * GET /manager/operations — where the bakery's work stands now: the orders
 * still to make or hand over by status, each station's tickets (a station
 * with none left out), today's pickups, the orders whose pickup slot ended
 * uncollected (the ten waiting longest, and how many in all), and how long the
 * orders that became ready today took.
 */
export const operationsSchema = z.object({
  orders: z.object({ pending: count, confirmed: count, in_production: count, ready: count }),
  stations: z.array(z.object({ station: stationSchema, queued: count, in_progress: count, ready: count })),
  pickups: z.object({ due: count, collected: count, to_collect: count, missed: count }),
  late: z.object({
    total: count,
    orders: z.array(z.object({ code: orderCodeSchema, status: orderStatusSchema, pickup_at: apiDateTimeSchema })),
  }),
  production: salesReportSchema.shape.production,
});

/** An engagement feature a customer may use — mirrors the backend engagement report. */
const engagementFeatureSchema = z.enum(["reviews", "favorites", "wishlist", "messages", "promotions", "assistant"]);

/**
 * GET /manager/engagement — the open customer accounts, how many used none,
 * one, or two or more of the engagement features, and how many used each.
 */
export const engagementReportSchema = z.object({
  customers: z.number().int().min(0),
  used_none: z.number().int().min(0),
  used_one: z.number().int().min(0),
  used_two_or_more: z.number().int().min(0),
  features: z.array(z.object({ feature: engagementFeatureSchema, customers: z.number().int().min(0) })),
});

/** The most characters a promotion's subject and message hold — mirrors backend `promotion.MaxSubjectLength` and `MaxBodyLength` (code points). */
export const PROMOTION_SUBJECT_MAX_LENGTH = 120;
export const PROMOTION_BODY_MAX_LENGTH = 5000;

/** GET /manager/promotions — a promotion sent; how many it went to is null while it is sending. */
export const managerPromotionSchema = z.object({
  id: z.uuid(),
  subject: z.string().min(1),
  status: z.enum(["sending", "sent"]),
  recipient_count: z.number().int().min(0).nullable(),
  created_at: apiDateTimeSchema,
  sender_name: z.string().nullable(),
});

export const promotionListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
});

/** GET /manager/promotions/audience — how many customers a promotion sent now goes to. */
export const promotionAudienceSchema = z.object({
  recipients: z.number().int().min(0),
});

/** A promotion as written; the API trims both and checks their characters again. */
export const promotionFormSchema = z.object({
  subject: z
    .string()
    .trim()
    .min(1, "Write a subject")
    .refine(
      (subject) => [...subject].length <= PROMOTION_SUBJECT_MAX_LENGTH,
      `A subject holds up to ${PROMOTION_SUBJECT_MAX_LENGTH} characters`,
    ),
  body: z
    .string()
    .trim()
    .min(1, "Write the message")
    .refine(
      (body) => [...body].length <= PROMOTION_BODY_MAX_LENGTH,
      `A message holds up to ${PROMOTION_BODY_MAX_LENGTH} characters`,
    ),
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

export type EngagementFeature = z.infer<typeof engagementFeatureSchema>;
export type SalesReportFilterInput = z.infer<typeof salesReportFilterSchema>;
export type SalesReport = z.infer<typeof salesReportSchema>;
export type Operations = z.infer<typeof operationsSchema>;
export type EngagementReport = z.infer<typeof engagementReportSchema>;
export type ManagerIncident = z.infer<typeof managerIncidentSchema>;
export type IncidentListFilterInput = z.input<typeof incidentListFilterSchema>;
export type IncidentSummary = z.infer<typeof incidentSummarySchema>;
export type IncidentsListResult = {
  incidents: ManagerIncident[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
};

export type ManagerPromotion = z.infer<typeof managerPromotionSchema>;
export type PromotionListFilterInput = z.infer<typeof promotionListFilterSchema>;
export type PromotionAudience = z.infer<typeof promotionAudienceSchema>;
export type PromotionFormInput = z.infer<typeof promotionFormSchema>;
export type PromotionsListResult = {
  promotions: ManagerPromotion[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
};
export type ManagerReview = z.infer<typeof managerReviewSchema>;
export type ReviewListFilterInput = z.infer<typeof reviewListFilterSchema>;
export type ReviewSummary = z.infer<typeof reviewSummarySchema>;

export type ReviewsListResult = {
  reviews: ManagerReview[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
};

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

/** The order code a manager looks an order up by, as typed. */
export const managerOrderLookupSchema = z.object({
  code: z
    .string()
    .trim()
    .toUpperCase()
    .refine((value) => orderCodeSchema.safeParse(value).success, "Enter the order code, such as CH-261125-001."),
});

export type ManagerOrderLookup = z.infer<typeof managerOrderLookupSchema>;

/** `POST /manager/orders/:id/handover` — how the manager checked who is collecting an order without its code. */
export const managerHandoverInputSchema = z.object({
  reason: reasonSchema("Say how you checked who is collecting it"),
});

export type ManagerHandoverInput = z.infer<typeof managerHandoverInputSchema>;

/** GET/PATCH /api/v1/manager/settings — the pickup rules a manager edits. */
export const storeSettingsSchema = z.object({
  opens_at: clockTimeSchema,
  closes_at: clockTimeSchema,
  preorder_min_lead_minutes: z.number().int().min(0),
  max_advance_days: z.number().int().min(1),
  slot_minutes: z.number().int().min(1),
  slot_capacity: z.number().int().min(1),
  instant_prep_minutes: z.number().int().min(0),
  payment_hold_minutes: z.number().int().min(1),
  updated_at: apiDateTimeSchema,
});

/**
 * Bounds of backend `store.Settings`: at most seven days' notice, 1 to 90 days
 * ahead, slots that divide an hour, 1 to 200 orders a slot, at most four
 * hours to pack an instant order, and 5 to 120 minutes to pay.
 */
export const MAX_PREORDER_LEAD_MINUTES = 7 * 24 * 60;
export const MIN_ADVANCE_DAYS = 1;
export const MAX_ADVANCE_DAYS = 90;
export const SLOT_MINUTE_OPTIONS = [10, 15, 20, 30, 60] as const;
export const MAX_SLOT_CAPACITY = 200;
export const MAX_INSTANT_PREP_MINUTES = 4 * 60;
export const MIN_PAYMENT_HOLD_MINUTES = 5;
export const MAX_PAYMENT_HOLD_MINUTES = 120;

const storeSettingsFieldsSchema = z.object({
  opens_at: clockTimeSchema,
  closes_at: clockTimeSchema,
  preorder_min_lead_minutes: z
    .number({ error: "Enter the notice in minutes" })
    .int()
    .min(0)
    .max(MAX_PREORDER_LEAD_MINUTES, `At most ${MAX_PREORDER_LEAD_MINUTES} minutes (7 days)`),
  max_advance_days: z
    .number({ error: "Enter the number of days" })
    .int()
    .min(MIN_ADVANCE_DAYS, `At least ${MIN_ADVANCE_DAYS} day`)
    .max(MAX_ADVANCE_DAYS, `At most ${MAX_ADVANCE_DAYS} days`),
  slot_minutes: z
    .number({ error: "Choose a slot length" })
    .refine((minutes) => SLOT_MINUTE_OPTIONS.some((option) => option === minutes), "Choose a slot length"),
  slot_capacity: z
    .number({ error: "Enter the number of orders" })
    .int()
    .min(1, "At least 1 order")
    .max(MAX_SLOT_CAPACITY, `At most ${MAX_SLOT_CAPACITY} orders`),
  instant_prep_minutes: z
    .number({ error: "Enter the time in minutes" })
    .int()
    .min(0)
    .max(MAX_INSTANT_PREP_MINUTES, `At most ${MAX_INSTANT_PREP_MINUTES} minutes (4 hours)`),
  payment_hold_minutes: z
    .number({ error: "Enter the time in minutes" })
    .int()
    .min(MIN_PAYMENT_HOLD_MINUTES, `At least ${MIN_PAYMENT_HOLD_MINUTES} minutes`)
    .max(MAX_PAYMENT_HOLD_MINUTES, `At most ${MAX_PAYMENT_HOLD_MINUTES} minutes (2 hours)`),
});

/**
 * PATCH /api/v1/manager/settings — only the fields the manager changed, so an edit
 * never overwrites a setting someone else saved meanwhile. The server checks
 * the merged result.
 */
export const storeSettingsPatchSchema = storeSettingsFieldsSchema.partial();

/** The settings form; mirrors backend `store.Settings.Validate`. */
export const storeSettingsFormSchema = storeSettingsFieldsSchema
  .refine((v) => clockToMinutes(v.opens_at) < clockToMinutes(v.closes_at), {
    path: ["closes_at"],
    message: "Closing time must be after opening time",
  })
  .refine((v) => v.preorder_min_lead_minutes < v.max_advance_days * 24 * 60, {
    path: ["preorder_min_lead_minutes"],
    message: "Notice must be shorter than the booking window",
  })
  // Only once the hours are in order: an inverted pair is one mistake, already named.
  .refine((v) => clockToMinutes(v.closes_at) <= clockToMinutes(v.opens_at) ||
    clockToMinutes(v.closes_at) - clockToMinutes(v.opens_at) >= v.slot_minutes, {
    path: ["slot_minutes"],
    message: "The opening hours must hold at least one slot",
  });

/** GET /api/v1/manager/closed-dates — a day the bakery takes no pickups. */
export const closedDateSchema = z.object({
  id: z.uuid(),
  date: z.iso.date(),
  reason: z.string().min(1),
  created_at: apiDateTimeSchema,
});

export const closedDateFormSchema = z.object({
  date: z.iso.date("Choose a day"),
  reason: reasonSchema("Tell customers why the bakery is closed"),
});

/**
 * A change that can strand booked pickups: a day the bakery closes, or new
 * opening and closing times from now on.
 */
export const strandingChangeSchema = z.union([
  z.object({ closed_date: z.iso.date() }),
  z.object({ opens_at: clockTimeSchema, closes_at: clockTimeSchema }),
]);

/** GET /api/v1/manager/settings/stranded-pickups — an order a change would strand. */
const strandedPickupSchema = z.object({
  id: z.uuid(),
  code: orderCodeSchema,
  status: orderStatusSchema,
  pickup_at: apiDateTimeSchema,
  customer: z.string().nullable(),
});

export const strandedPickupsSchema = z.object({
  orders: z.array(strandedPickupSchema),
  total: z.number().int().min(0),
});

/** The reason the stranded orders' customers read. */
export const cancelStrandedFormSchema = z.object({
  reason: reasonSchema("Tell the customers why"),
});

export const cancelStrandedResultSchema = z.object({
  cancelled: z.number().int().min(0),
});

export type StrandingChange = z.infer<typeof strandingChangeSchema>;
export type StrandedPickups = z.infer<typeof strandedPickupsSchema>;
export type CancelStrandedFormInput = z.infer<typeof cancelStrandedFormSchema>;
export type StoreSettings = z.infer<typeof storeSettingsSchema>;
export type StoreSettingsFormInput = z.infer<typeof storeSettingsFormSchema>;
export type StoreSettingsPatch = z.infer<typeof storeSettingsPatchSchema>;
export type ClosedDate = z.infer<typeof closedDateSchema>;
export type ClosedDateFormInput = z.infer<typeof closedDateFormSchema>;
