"use client";

import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { toast } from "sonner";

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
  getProductById,
  listCategories,
  listCombos,
  listDiscountCodes,
  listProducts,
  updateCategory,
  updateCombo,
  updateDiscountCode,
  updateProduct,
} from "../api";
import {
  categoryListFilterSchema,
  comboListFilterSchema,
  discountCodeListFilterSchema,
  productListFilterSchema,
  type CategoryFormInput,
  type CategoryListFilterInput,
  type ComboFormInput,
  type ComboListFilterInput,
  type DiscountCodeFormInput,
  type DiscountCodeListFilterInput,
  type ProductFormInput,
  type ProductListFilterInput,
} from "../schemas";
import { managerQueryKeys } from "./query-options";

export { managerQueryKeys } from "./query-options";

export function useCategories(input: CategoryListFilterInput) {
  const filter = categoryListFilterSchema.parse(input);
  return useQuery({
    queryKey: managerQueryKeys.categories(filter),
    queryFn: () => listCategories(filter),
    placeholderData: keepPreviousData,
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
