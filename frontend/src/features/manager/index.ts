/**
 * Manager feature — catalog CRUD (manager-only).
 *
 * Internal: api/, components/, hooks/, schemas/
 */
export {
  CategoryForm,
  ComboForm,
  DiscountCodeForm,
  ManagerCategoriesTable,
  ManagerCombosTable,
  ManagerDiscountCodesTable,
  ManagerProductsTable,
  ProductForm,
} from "./components";
export {
  managerCategoriesBreadcrumb,
  managerCombosBreadcrumb,
  managerDiscountCodesBreadcrumb,
  managerProductsBreadcrumb,
} from "./lib/manager-breadcrumbs";
export {
  useCategories,
  useCategory,
  useCombo,
  useCombos,
  useCreateCategory,
  useCreateCombo,
  useCreateDiscountCode,
  useCreateProduct,
  useDeleteCategory,
  useDeleteCombo,
  useDeleteDiscountCode,
  useDeleteProduct,
  useDiscountCode,
  useDiscountCodes,
  useProduct,
  useProducts,
  useUpdateCategory,
  useUpdateCombo,
  useUpdateDiscountCode,
  useUpdateProduct,
} from "./hooks";
