import { z } from "zod";

import {
  browserRequest,
  browserRequestVoid,
  browserRequestWithMeta,
} from "@/lib/browser-api-client";
import { parsePaginatedList } from "@/lib/pagination/parse-paginated-list";

import {
  categoryFormSchema,
  categoryListFilterSchema,
  managerCategorySchema,
  managerProductSchema,
  comboFormSchema,
  comboListFilterSchema,
  discountCodeFormSchema,
  discountCodeListFilterSchema,
  incidentListFilterSchema,
  incidentSummarySchema,
  managerComboSchema,
  managerIncidentSchema,
  managerDiscountCodeSchema,
  managerPromotionSchema,
  managerReviewSchema,
  productFormSchema,
  productListFilterSchema,
  promotionAudienceSchema,
  promotionFormSchema,
  promotionListFilterSchema,
  reviewListFilterSchema,
  reviewModerationSchema,
  reviewSummarySchema,
  type CategoriesListResult,
  type CategoryFormInput,
  type CategoryListFilterInput,
  type ComboFormInput,
  type ComboListFilterInput,
  type CombosListResult,
  type DiscountCodeFormInput,
  type DiscountCodeListFilterInput,
  type DiscountCodesListResult,
  type IncidentListFilterInput,
  type IncidentSummary,
  type IncidentsListResult,
  type ManagerIncident,
  type ManagerCategory,
  type ManagerCombo,
  type ManagerDiscountCode,
  type ManagerProduct,
  type ManagerPromotion,
  type ManagerReview,
  type ProductFormInput,
  type ProductListFilterInput,
  type ProductsListResult,
  type PromotionAudience,
  type PromotionFormInput,
  type PromotionListFilterInput,
  type PromotionsListResult,
  type ReviewListFilterInput,
  type ReviewModeration,
  type ReviewsListResult,
  type ReviewSummary,
} from "../schemas";

function categoriesPath(filter: CategoryListFilterInput): string {
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  if (filter.search) {
    params.set("search", filter.search);
  }
  return `/api/v1/manager/categories?${params.toString()}`;
}

function productsPath(filter: ProductListFilterInput): string {
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  if (filter.search) {
    params.set("search", filter.search);
  }
  if (filter.category_id) {
    params.set("category_id", filter.category_id);
  }
  return `/api/v1/manager/products?${params.toString()}`;
}

export async function listCategories(
  input: CategoryListFilterInput,
): Promise<CategoriesListResult> {
  const filter = categoryListFilterSchema.parse(input);
  const result = await browserRequestWithMeta<ManagerCategory[]>(
    categoriesPath(filter),
    { method: "GET", schema: z.array(managerCategorySchema) },
  );
  const parsed = parsePaginatedList(result.data, result.meta, {
    page: filter.page,
    page_size: filter.page_size,
  });
  return {
    categories: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}

export async function getCategoryById(id: string): Promise<ManagerCategory> {
  const parsedId = z.uuid().parse(id);
  return browserRequest<ManagerCategory>(`/api/v1/manager/categories/${parsedId}`, {
    method: "GET",
    schema: managerCategorySchema,
  });
}

export async function createCategory(
  input: CategoryFormInput,
): Promise<ManagerCategory> {
  const body = categoryFormSchema.parse(input);
  return browserRequest<ManagerCategory>("/api/v1/manager/categories", {
    method: "POST",
    json: body,
    schema: managerCategorySchema,
  });
}

export async function updateCategory(
  id: string,
  input: CategoryFormInput,
): Promise<ManagerCategory> {
  const parsedId = z.uuid().parse(id);
  const body = categoryFormSchema.parse(input);
  return browserRequest<ManagerCategory>(`/api/v1/manager/categories/${parsedId}`, {
    method: "PATCH",
    json: body,
    schema: managerCategorySchema,
  });
}

export async function deleteCategory(id: string): Promise<void> {
  const parsedId = z.uuid().parse(id);
  await browserRequestVoid(`/api/v1/manager/categories/${parsedId}`, {
    method: "DELETE",
  });
}

export async function listProducts(
  input: ProductListFilterInput,
): Promise<ProductsListResult> {
  const filter = productListFilterSchema.parse(input);
  const result = await browserRequestWithMeta<ManagerProduct[]>(
    productsPath(filter),
    { method: "GET", schema: z.array(managerProductSchema) },
  );
  const parsed = parsePaginatedList(result.data, result.meta, {
    page: filter.page,
    page_size: filter.page_size,
  });
  return {
    products: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}

export async function getProductById(id: string): Promise<ManagerProduct> {
  const parsedId = z.uuid().parse(id);
  return browserRequest<ManagerProduct>(`/api/v1/manager/products/${parsedId}`, {
    method: "GET",
    schema: managerProductSchema,
  });
}

export async function createProduct(
  input: ProductFormInput,
): Promise<ManagerProduct> {
  const body = productFormSchema.parse(input);
  return browserRequest<ManagerProduct>("/api/v1/manager/products", {
    method: "POST",
    json: body,
    schema: managerProductSchema,
  });
}

export async function updateProduct(
  id: string,
  input: ProductFormInput,
): Promise<ManagerProduct> {
  const parsedId = z.uuid().parse(id);
  const body = productFormSchema.parse(input);
  return browserRequest<ManagerProduct>(`/api/v1/manager/products/${parsedId}`, {
    method: "PATCH",
    json: body,
    schema: managerProductSchema,
  });
}

export async function deleteProduct(id: string): Promise<void> {
  const parsedId = z.uuid().parse(id);
  await browserRequestVoid(`/api/v1/manager/products/${parsedId}`, {
    method: "DELETE",
  });
}

function combosPath(filter: ComboListFilterInput): string {
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  if (filter.search) {
    params.set("search", filter.search);
  }
  return `/api/v1/manager/combos?${params.toString()}`;
}

function discountCodesPath(filter: DiscountCodeListFilterInput): string {
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  if (filter.search) {
    params.set("search", filter.search);
  }
  return `/api/v1/manager/discount-codes?${params.toString()}`;
}

export async function listCombos(
  input: ComboListFilterInput,
): Promise<CombosListResult> {
  const filter = comboListFilterSchema.parse(input);
  const result = await browserRequestWithMeta<ManagerCombo[]>(
    combosPath(filter),
    { method: "GET", schema: z.array(managerComboSchema) },
  );
  const parsed = parsePaginatedList(result.data, result.meta, {
    page: filter.page,
    page_size: filter.page_size,
  });
  return {
    combos: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}

export async function getComboById(id: string): Promise<ManagerCombo> {
  const parsedId = z.uuid().parse(id);
  return browserRequest<ManagerCombo>(`/api/v1/manager/combos/${parsedId}`, {
    method: "GET",
    schema: managerComboSchema,
  });
}

export async function createCombo(input: ComboFormInput): Promise<ManagerCombo> {
  const body = comboFormSchema.parse(input);
  return browserRequest<ManagerCombo>("/api/v1/manager/combos", {
    method: "POST",
    json: body,
    schema: managerComboSchema,
  });
}

export async function updateCombo(
  id: string,
  input: ComboFormInput,
): Promise<ManagerCombo> {
  const parsedId = z.uuid().parse(id);
  const body = comboFormSchema.parse(input);
  return browserRequest<ManagerCombo>(`/api/v1/manager/combos/${parsedId}`, {
    method: "PATCH",
    json: body,
    schema: managerComboSchema,
  });
}

export async function deleteCombo(id: string): Promise<void> {
  const parsedId = z.uuid().parse(id);
  await browserRequestVoid(`/api/v1/manager/combos/${parsedId}`, {
    method: "DELETE",
  });
}

export async function listDiscountCodes(
  input: DiscountCodeListFilterInput,
): Promise<DiscountCodesListResult> {
  const filter = discountCodeListFilterSchema.parse(input);
  const result = await browserRequestWithMeta<ManagerDiscountCode[]>(
    discountCodesPath(filter),
    { method: "GET", schema: z.array(managerDiscountCodeSchema) },
  );
  const parsed = parsePaginatedList(result.data, result.meta, {
    page: filter.page,
    page_size: filter.page_size,
  });
  return {
    discount_codes: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}

export async function getDiscountCodeById(
  id: string,
): Promise<ManagerDiscountCode> {
  const parsedId = z.uuid().parse(id);
  return browserRequest<ManagerDiscountCode>(
    `/api/v1/manager/discount-codes/${parsedId}`,
    {
      method: "GET",
      schema: managerDiscountCodeSchema,
    },
  );
}

export async function createDiscountCode(
  input: DiscountCodeFormInput,
): Promise<ManagerDiscountCode> {
  const body = discountCodeFormSchema.parse(input);
  return browserRequest<ManagerDiscountCode>("/api/v1/manager/discount-codes", {
    method: "POST",
    json: body,
    schema: managerDiscountCodeSchema,
  });
}

export async function updateDiscountCode(
  id: string,
  input: DiscountCodeFormInput,
): Promise<ManagerDiscountCode> {
  const parsedId = z.uuid().parse(id);
  const body = discountCodeFormSchema.parse(input);
  return browserRequest<ManagerDiscountCode>(
    `/api/v1/manager/discount-codes/${parsedId}`,
    {
      method: "PATCH",
      json: body,
      schema: managerDiscountCodeSchema,
    },
  );
}

export async function deleteDiscountCode(id: string): Promise<void> {
  const parsedId = z.uuid().parse(id);
  await browserRequestVoid(`/api/v1/manager/discount-codes/${parsedId}`, {
    method: "DELETE",
  });
}

/** A page of reviews, latest first, narrowed to one status when one is given. */
export async function listReviews(input: ReviewListFilterInput): Promise<ReviewsListResult> {
  const filter = reviewListFilterSchema.parse(input);
  const params = new URLSearchParams({ page: String(filter.page), page_size: String(filter.page_size) });
  if (filter.status) {
    params.set("status", filter.status);
  }
  const result = await browserRequestWithMeta<ManagerReview[]>(`/api/v1/manager/reviews?${params.toString()}`, {
    method: "GET",
    schema: z.array(managerReviewSchema),
  });
  const parsed = parsePaginatedList(result.data, result.meta, { page: filter.page, page_size: filter.page_size });
  return { reviews: parsed.items, pagination: parsed.pagination };
}

/** Every review not hidden, added up. */
export async function getReviewSummary(): Promise<ReviewSummary> {
  return browserRequest<ReviewSummary>("/api/v1/manager/reviews/summary", {
    method: "GET",
    schema: reviewSummarySchema,
  });
}

/** Publishes a review on the storefront, or hides it. */
export async function moderateReview(id: string, status: ReviewModeration): Promise<ManagerReview> {
  const parsedId = z.uuid().parse(id);
  return browserRequest<ManagerReview>(`/api/v1/manager/reviews/${parsedId}`, {
    method: "PATCH",
    schema: managerReviewSchema,
    json: { status: reviewModerationSchema.parse(status) },
  });
}

/** A page of a week's incidents, latest first, narrowed to one type when one is given. */
export async function listIncidents(input: IncidentListFilterInput): Promise<IncidentsListResult> {
  const filter = incidentListFilterSchema.parse(input);
  const params = new URLSearchParams({ week: filter.week, page: String(filter.page), page_size: String(filter.page_size) });
  if (filter.type) {
    params.set("type", filter.type);
  }
  const result = await browserRequestWithMeta<ManagerIncident[]>(`/api/v1/manager/incidents?${params.toString()}`, {
    method: "GET",
    schema: z.array(managerIncidentSchema),
  });
  const parsed = parsePaginatedList(result.data, result.meta, { page: filter.page, page_size: filter.page_size });
  return { incidents: parsed.items, pagination: parsed.pagination };
}

/** How many incidents of each type the week starting on the Monday `week` holds. */
export async function getIncidentSummary(week: string): Promise<IncidentSummary> {
  const params = new URLSearchParams({ week: z.iso.date().parse(week) });
  return browserRequest<IncidentSummary>(`/api/v1/manager/incidents/summary?${params.toString()}`, {
    method: "GET",
    schema: incidentSummarySchema,
  });
}

/** A page of the promotions sent, latest first. */
export async function listPromotions(input: PromotionListFilterInput): Promise<PromotionsListResult> {
  const filter = promotionListFilterSchema.parse(input);
  const params = new URLSearchParams({ page: String(filter.page), page_size: String(filter.page_size) });
  const result = await browserRequestWithMeta<ManagerPromotion[]>(`/api/v1/manager/promotions?${params.toString()}`, {
    method: "GET",
    schema: z.array(managerPromotionSchema),
  });
  const parsed = parsePaginatedList(result.data, result.meta, { page: filter.page, page_size: filter.page_size });
  return { promotions: parsed.items, pagination: parsed.pagination };
}

/** How many customers a promotion sent now goes to. */
export async function getPromotionAudience(): Promise<PromotionAudience> {
  return browserRequest<PromotionAudience>("/api/v1/manager/promotions/audience", {
    method: "GET",
    schema: promotionAudienceSchema,
  });
}

/** Sends a promotion: the worker emails it to every customer who agreed to promotions. */
export async function sendPromotion(input: PromotionFormInput): Promise<ManagerPromotion> {
  return browserRequest<ManagerPromotion>("/api/v1/manager/promotions", {
    method: "POST",
    schema: managerPromotionSchema,
    json: promotionFormSchema.parse(input),
  });
}
