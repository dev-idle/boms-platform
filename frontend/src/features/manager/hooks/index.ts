"use client";

import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { toast } from "sonner";

import { isApiError } from "@/lib/errors";

import {
  createCategory,
  createCombo,
  createDiscountCode,
  createProduct,
  deleteCategory,
  deleteCombo,
  deleteDiscountCode,
  deleteProduct,
  getCategoryById,
  getComboById,
  getDiscountCodeById,
  getEngagementReport,
  getIncidentSummary,
  getProductById,
  getPromotionAudience,
  getReviewSummary,
  getSalesReport,
  listCategories,
  listCombos,
  listDiscountCodes,
  listIncidents,
  listProducts,
  listPromotions,
  listReviews,
  moderateReview,
  sendPromotion,
  updateCategory,
  updateCombo,
  updateDiscountCode,
  updateProduct,
} from "../api";
import {
  categoryListFilterSchema,
  comboListFilterSchema,
  discountCodeListFilterSchema,
  incidentListFilterSchema,
  productListFilterSchema,
  promotionListFilterSchema,
  reviewListFilterSchema,
  salesReportFilterSchema,
  type CategoryFormInput,
  type CategoryListFilterInput,
  type ComboFormInput,
  type ComboListFilterInput,
  type DiscountCodeFormInput,
  type DiscountCodeListFilterInput,
  type IncidentListFilterInput,
  type ProductFormInput,
  type ProductListFilterInput,
  type PromotionFormInput,
  type PromotionListFilterInput,
  type ReviewListFilterInput,
  type ReviewModeration,
  type SalesReportFilterInput,
} from "../schemas";
import { managerQueryKeys, MANAGER_LIST_STALE_TIME_MS } from "./query-options";

export function useCategories(input: CategoryListFilterInput) {
  const filter = categoryListFilterSchema.parse(input);
  return useQuery({
    queryKey: managerQueryKeys.categories(filter),
    queryFn: () => listCategories(filter),
    placeholderData: keepPreviousData,
    staleTime: MANAGER_LIST_STALE_TIME_MS,
  });
}

export function useCategory(id: string) {
  return useQuery({
    queryKey: managerQueryKeys.category(id),
    queryFn: () => getCategoryById(id),
    enabled: Boolean(id),
  });
}

export function useCreateCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CategoryFormInput) => createCategory(input),
    onSuccess: () => {
      toast.success("Category created");
      void queryClient.invalidateQueries({ queryKey: managerQueryKeys.categoriesRoot });
    },
  });
}

export function useUpdateCategory(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CategoryFormInput) => updateCategory(id, input),
    onSuccess: (category) => {
      toast.success("Category updated");
      queryClient.setQueryData(managerQueryKeys.category(id), category);
      void queryClient.invalidateQueries({ queryKey: managerQueryKeys.categoriesRoot });
    },
  });
}

export function useDeleteCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteCategory(id),
    onSuccess: () => {
      toast.success("Category deleted");
      void queryClient.invalidateQueries({ queryKey: managerQueryKeys.categoriesRoot });
    },
  });
}

export function useProducts(input: ProductListFilterInput) {
  const filter = productListFilterSchema.parse(input);
  return useQuery({
    queryKey: managerQueryKeys.products(filter),
    queryFn: () => listProducts(filter),
    placeholderData: keepPreviousData,
    staleTime: MANAGER_LIST_STALE_TIME_MS,
  });
}

export function useProduct(id: string) {
  return useQuery({
    queryKey: managerQueryKeys.product(id),
    queryFn: () => getProductById(id),
    enabled: Boolean(id),
  });
}

export function useCreateProduct() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: ProductFormInput) => createProduct(input),
    onSuccess: () => {
      toast.success("Product created");
      void queryClient.invalidateQueries({ queryKey: managerQueryKeys.productsRoot });
    },
  });
}

export function useUpdateProduct(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: ProductFormInput) => updateProduct(id, input),
    onSuccess: (product) => {
      toast.success("Product updated");
      queryClient.setQueryData(managerQueryKeys.product(id), product);
      void queryClient.invalidateQueries({ queryKey: managerQueryKeys.productsRoot });
    },
  });
}

export function useDeleteProduct() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteProduct(id),
    onSuccess: () => {
      toast.success("Product deleted");
      void queryClient.invalidateQueries({ queryKey: managerQueryKeys.productsRoot });
    },
  });
}

export function useCombos(input: ComboListFilterInput) {
  const filter = comboListFilterSchema.parse(input);
  return useQuery({
    queryKey: managerQueryKeys.combos(filter),
    queryFn: () => listCombos(filter),
    placeholderData: keepPreviousData,
    staleTime: MANAGER_LIST_STALE_TIME_MS,
  });
}

export function useCombo(id: string) {
  return useQuery({
    queryKey: managerQueryKeys.combo(id),
    queryFn: () => getComboById(id),
    enabled: Boolean(id),
  });
}

export function useCreateCombo() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: ComboFormInput) => createCombo(input),
    onSuccess: () => {
      toast.success("Combo created");
      void queryClient.invalidateQueries({ queryKey: managerQueryKeys.combosRoot });
    },
  });
}

export function useUpdateCombo(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: ComboFormInput) => updateCombo(id, input),
    onSuccess: (combo) => {
      toast.success("Combo updated");
      queryClient.setQueryData(managerQueryKeys.combo(id), combo);
      void queryClient.invalidateQueries({ queryKey: managerQueryKeys.combosRoot });
    },
  });
}

export function useDeleteCombo() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteCombo(id),
    onSuccess: () => {
      toast.success("Combo deleted");
      void queryClient.invalidateQueries({ queryKey: managerQueryKeys.combosRoot });
    },
  });
}

export function useDiscountCodes(input: DiscountCodeListFilterInput) {
  const filter = discountCodeListFilterSchema.parse(input);
  return useQuery({
    queryKey: managerQueryKeys.discountCodes(filter),
    queryFn: () => listDiscountCodes(filter),
    placeholderData: keepPreviousData,
    staleTime: MANAGER_LIST_STALE_TIME_MS,
  });
}

export function useDiscountCode(id: string) {
  return useQuery({
    queryKey: managerQueryKeys.discountCode(id),
    queryFn: () => getDiscountCodeById(id),
    enabled: Boolean(id),
  });
}

export function useCreateDiscountCode() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: DiscountCodeFormInput) => createDiscountCode(input),
    onSuccess: () => {
      toast.success("Discount code created");
      void queryClient.invalidateQueries({
        queryKey: managerQueryKeys.discountCodesRoot,
      });
    },
  });
}

export function useUpdateDiscountCode(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: DiscountCodeFormInput) => updateDiscountCode(id, input),
    onSuccess: (discountCode) => {
      toast.success("Discount code updated");
      queryClient.setQueryData(managerQueryKeys.discountCode(id), discountCode);
      void queryClient.invalidateQueries({
        queryKey: managerQueryKeys.discountCodesRoot,
      });
    },
  });
}

export function useDeleteDiscountCode() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteDiscountCode(id),
    onSuccess: () => {
      toast.success("Discount code deleted");
      void queryClient.invalidateQueries({
        queryKey: managerQueryKeys.discountCodesRoot,
      });
    },
  });
}

/** A page of reviews to moderate; refreshed live as customers write them. */
export function useReviews(input: ReviewListFilterInput) {
  const filter = reviewListFilterSchema.parse(input);
  return useQuery({
    queryKey: managerQueryKeys.reviews(filter),
    queryFn: () => listReviews(filter),
    placeholderData: keepPreviousData,
    staleTime: MANAGER_LIST_STALE_TIME_MS,
  });
}

/** Every review not hidden, added up. */
export function useReviewSummary() {
  return useQuery({
    queryKey: managerQueryKeys.reviewSummary,
    queryFn: getReviewSummary,
    staleTime: MANAGER_LIST_STALE_TIME_MS,
  });
}

/** Publishes a review or hides it; the table holds several, so the review rides with each call. */
export function useModerateReview() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: ReviewModeration }) => moderateReview(id, status),
    onSuccess: (review) => {
      void queryClient.invalidateQueries({ queryKey: managerQueryKeys.reviewsRoot });
      toast.success(review.status === "published" ? "Review published" : "Review hidden");
    },
    onError: (error) => {
      toast.error(isApiError(error) ? error.message : "Failed to update the review");
    },
  });
}

/** What the bakery sold over a range; a new range loads afresh, so no figure of the last one stays on screen. */
export function useSalesReport(input: SalesReportFilterInput) {
  const filter = salesReportFilterSchema.parse(input);
  return useQuery({
    queryKey: managerQueryKeys.salesReport(filter),
    queryFn: () => getSalesReport(filter),
    staleTime: MANAGER_LIST_STALE_TIME_MS,
  });
}

/** How customers use the engagement features. */
export function useEngagementReport() {
  return useQuery({
    queryKey: managerQueryKeys.engagementReport,
    queryFn: getEngagementReport,
    staleTime: MANAGER_LIST_STALE_TIME_MS,
  });
}

/** A page of a week's incidents; refreshed live as they are recorded. */
export function useIncidents(input: IncidentListFilterInput) {
  const filter = incidentListFilterSchema.parse(input);
  return useQuery({
    queryKey: managerQueryKeys.incidents(filter),
    queryFn: () => listIncidents(filter),
    placeholderData: keepPreviousData,
    staleTime: MANAGER_LIST_STALE_TIME_MS,
  });
}

/** How many incidents of each type a week holds. */
export function useIncidentSummary(week: string) {
  return useQuery({
    queryKey: managerQueryKeys.incidentSummary(week),
    queryFn: () => getIncidentSummary(week),
    staleTime: MANAGER_LIST_STALE_TIME_MS,
  });
}

/** A page of the promotions sent; refreshed live as they go out. */
export function usePromotions(input: PromotionListFilterInput) {
  const filter = promotionListFilterSchema.parse(input);
  return useQuery({
    queryKey: managerQueryKeys.promotions(filter),
    queryFn: () => listPromotions(filter),
    placeholderData: keepPreviousData,
    staleTime: MANAGER_LIST_STALE_TIME_MS,
  });
}

/** How many customers a promotion sent now goes to, read afresh each time the form opens. */
export function usePromotionAudience() {
  return useQuery({
    queryKey: managerQueryKeys.promotionAudience,
    queryFn: getPromotionAudience,
    staleTime: 0,
  });
}

/** Sends a promotion; the form places field errors itself. */
export function useSendPromotion() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: PromotionFormInput) => sendPromotion(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: managerQueryKeys.promotionsRoot });
      toast.success("Promotion on its way");
    },
  });
}
