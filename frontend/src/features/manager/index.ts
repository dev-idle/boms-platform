/**
 * Manager feature — catalog CRUD, review moderation, promotions, the incident log and the engagement report (manager-only).
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
  ManagerEngagement,
  ManagerIncidents,
  ManagerLiveUpdates,
  ManagerNewPromotion,
  ManagerProductsTable,
  ManagerPromotionsTable,
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
