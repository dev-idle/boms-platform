/**
 * User feature — self-service profile, password, and account lifecycle.
 *
 * Internal: api/, components/, hooks/, schemas/, types/
 * Session identity: re-exported via auth; `/api/v1/me` is the source of truth.
 */
export { changePassword, deleteAccount, getMe, updateProfile } from "./api";
export {
  AdminAccountProfileForm,
  ChangePasswordForm,
  CustomerAccountProfileForm,
  DeleteAccountCard,
  OperationalAccountProfileForm,
} from "./components";
export {
  meQueryOptions,
  useChangePassword,
  useDeleteAccount,
  useMe,
  useUpdateProfile,
  userQueryKeys,
} from "./hooks";
export {
  homeRouteForRole,
  isPathAllowedForRole,
  passwordRouteForRole,
  profileRouteForRole,
  routePrefixesForRole,
} from "@/lib/routing/role-routes";
export {
  adminProfileSchema,
  changePasswordFormSchema,
  changePasswordSchema,
  customerProfileSchema,
  meSchema,
  staffProfileSchema,
  updateSelfProfileSchema,
  type ChangePasswordFormInput,
  type ChangePasswordInput,
  type UpdateSelfProfileInput,
} from "./schemas/index";
export type { AdminProfile, CustomerProfile, Me, StaffProfile } from "./types";
