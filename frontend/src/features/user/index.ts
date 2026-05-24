/**
 * User feature — self-service profile, password, and account lifecycle.
 *
 * Internal: api/, components/, hooks/, lib/, schemas/, types/
 * Session identity: re-exported via auth; `/api/v1/me` is the source of truth.
 */
export { changePassword, deleteAccount, getMe, updateProfile } from "./api";
export {
  AdminAccountProfileForm,
  ChangePasswordForm,
  CustomerAccountProfileForm,
  DeleteAccountCard,
  StaffAccountProfileForm,
} from "./components";
export {
  useChangePassword,
  useDeleteAccount,
  useMe,
  useUpdateProfile,
  userQueryKeys,
} from "./hooks";
export { passwordRouteForRole, profileRouteForRole } from "./lib/role-routes";
export {
  adminProfileSchema,
  changePasswordSchema,
  customerProfileSchema,
  meSchema,
  staffProfileSchema,
  updateSelfProfileSchema,
  type ChangePasswordInput,
  type UpdateSelfProfileInput,
} from "./schemas/index";
export type { AdminProfile, CustomerProfile, Me, StaffProfile } from "./types";
