/**
 * Manager feature — catalog CRUD and review moderation (manager-only).
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
  ManagerLiveUpdates,
  ManagerProductsTable,
  ManagerReviews,
  ProductForm,
} from "./components";
export {
  managerCategoriesBreadcrumb,
  managerCombosBreadcrumb,
  managerDiscountCodesBreadcrumb,
  managerProductsBreadcrumb,
} from "./lib/manager-breadcrumbs";
export {
  useCategory,
  useCombo,
  useDiscountCode,
  useProduct,
} from "./hooks";
