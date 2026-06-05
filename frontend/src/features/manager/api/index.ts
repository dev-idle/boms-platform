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
  productFormSchema,
  productListFilterSchema,
  type CategoriesListResult,
  type CategoryFormInput,
  type CategoryListFilterInput,
  type ManagerCategory,
  type ManagerProduct,
  type ProductFormInput,
  type ProductListFilterInput,
  type ProductsListResult,
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
  const parsedId = z.string().uuid().parse(id);
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
  const parsedId = z.string().uuid().parse(id);
  const body = categoryFormSchema.parse(input);
  return browserRequest<ManagerCategory>(`/api/v1/manager/categories/${parsedId}`, {
    method: "PATCH",
    json: body,
    schema: managerCategorySchema,
  });
}

export async function deleteCategory(id: string): Promise<void> {
  const parsedId = z.string().uuid().parse(id);
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
  const parsedId = z.string().uuid().parse(id);
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
  const parsedId = z.string().uuid().parse(id);
  const body = productFormSchema.parse(input);
  return browserRequest<ManagerProduct>(`/api/v1/manager/products/${parsedId}`, {
    method: "PATCH",
    json: body,
    schema: managerProductSchema,
  });
}

export async function deleteProduct(id: string): Promise<void> {
  const parsedId = z.string().uuid().parse(id);
  await browserRequestVoid(`/api/v1/manager/products/${parsedId}`, {
    method: "DELETE",
  });
}
