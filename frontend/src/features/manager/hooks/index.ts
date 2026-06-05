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
  createProduct,
  deleteCategory,
  deleteProduct,
  getCategoryById,
  getProductById,
  listCategories,
  listProducts,
  updateCategory,
  updateProduct,
} from "../api";
import {
  categoryListFilterSchema,
  productListFilterSchema,
  type CategoryFormInput,
  type CategoryListFilterInput,
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
