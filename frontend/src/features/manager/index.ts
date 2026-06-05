/**
 * Manager feature — catalog CRUD (manager-only).
 *
 * Internal: api/, components/, hooks/, schemas/
 */
export {
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
} from "./api";
export {
  CategoryForm,
  ManagerCategoriesTable,
  ManagerProductsTable,
  ProductForm,
} from "./components";
export {
  managerQueryKeys,
  useCategories,
  useCategory,
  useCreateCategory,
  useCreateProduct,
  useDeleteCategory,
  useDeleteProduct,
  useProduct,
  useProducts,
  useUpdateCategory,
  useUpdateProduct,
} from "./hooks";
export {
  categoryFormSchema,
  categoryListFilterSchema,
  managerCategorySchema,
  managerProductSchema,
  productFormSchema,
  productListFilterSchema,
} from "./schemas";
export type {
  CategoriesListResult,
  CategoryFormInput,
  CategoryListFilterInput,
  ManagerCategory,
  ManagerProduct,
  ProductFormInput,
  ProductListFilterInput,
  ProductsListResult,
} from "./schemas";
