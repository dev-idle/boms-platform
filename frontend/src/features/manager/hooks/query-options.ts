import type { QueryKey } from "@tanstack/react-query";

import { REALTIME_EVENT_TYPE, type RealtimeEvent } from "@/lib/realtime/events";

import type {
  CategoryListFilterInput,
  ComboListFilterInput,
  DiscountCodeListFilterInput,
  ProductListFilterInput,
  PromotionListFilterInput,
  ReviewListFilterInput,
} from "../schemas";

/** List queries: align with global default; explicit for manager catalog tables. */
export const MANAGER_LIST_STALE_TIME_MS = 30_000;

export const managerQueryKeys = {
  all: ["manager"] as const,
  categoriesRoot: ["manager", "categories"] as const,
  categories: (filter: CategoryListFilterInput) =>
    [...managerQueryKeys.categoriesRoot, filter] as const,
  category: (id: string) => [...managerQueryKeys.all, "category", id] as const,
  productsRoot: ["manager", "products"] as const,
  products: (filter: ProductListFilterInput) =>
    [...managerQueryKeys.productsRoot, filter] as const,
  product: (id: string) => [...managerQueryKeys.all, "product", id] as const,
  combosRoot: ["manager", "combos"] as const,
  combos: (filter: ComboListFilterInput) =>
    [...managerQueryKeys.combosRoot, filter] as const,
  combo: (id: string) => [...managerQueryKeys.all, "combo", id] as const,
  discountCodesRoot: ["manager", "discount-codes"] as const,
  discountCodes: (filter: DiscountCodeListFilterInput) =>
    [...managerQueryKeys.discountCodesRoot, filter] as const,
  discountCode: (id: string) =>
    [...managerQueryKeys.all, "discount-code", id] as const,
  reviewsRoot: ["manager", "reviews"] as const,
  reviews: (filter: ReviewListFilterInput) =>
    [...managerQueryKeys.reviewsRoot, filter] as const,
  reviewSummary: ["manager", "reviews", "summary"] as const,
  promotionsRoot: ["manager", "promotions"] as const,
  promotions: (filter: PromotionListFilterInput) =>
    [...managerQueryKeys.promotionsRoot, filter] as const,
  promotionAudience: ["manager", "promotions", "audience"] as const,
};

/** Everything pushed events can change in a manager tab, refetched after a gap. */
export const managerLiveQueryKeys: readonly QueryKey[] = [managerQueryKeys.reviewsRoot, managerQueryKeys.promotionsRoot];

/**
 * Queries a pushed event makes stale in a manager tab: a review written, or
 * published or hidden in another tab, changes the reviews and their summary; a
 * promotion sent, or done sending, changes the promotions.
 */
export function managerQueryKeysForEvent(event: RealtimeEvent): QueryKey[] {
  switch (event.type) {
    case REALTIME_EVENT_TYPE.reviewChanged:
      return [managerQueryKeys.reviewsRoot];
    case REALTIME_EVENT_TYPE.promotionCreated:
    case REALTIME_EVENT_TYPE.promotionSent:
      return [managerQueryKeys.promotionsRoot];
    default:
      return [];
  }
}
