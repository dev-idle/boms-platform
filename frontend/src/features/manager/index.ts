/**
 * Manager feature — the operations dashboard, catalog CRUD, review moderation, promotions, the incident log, the sales and engagement reports and the store's pickup settings (manager-only).
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
  ManagerOperations,
  ManagerProductsTable,
  ManagerPromotionsTable,
  ManagerReviews,
  ManagerSalesReport,
  ManagerStoreSettings,
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
