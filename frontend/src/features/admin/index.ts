/**
 * Admin feature — the admin home, operational user management and store
 * settings (admin-only).
 *
 * Internal: api/, components/, hooks/, lib/, schemas/
 */
export {
  AdminDashboard,
  AdminStoreSettings,
  AdminUserDetail,
  AdminUsersTable,
  CreateOperationalUserForm,
} from "./components";
export { adminUsersBreadcrumb } from "./lib/admin-breadcrumbs";
