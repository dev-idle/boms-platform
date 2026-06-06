/**
 * User feature — self-service profile, password, and account lifecycle.
 *
 * Internal: api/, components/, hooks/, schemas/, types/
 * Session identity: re-exported via auth; `/api/v1/me` is the source of truth.
 */
export { changePassword, deleteAccount, getMe, updateProfile } from "./api";
export {
  AdminAccountProfileForm,
  AdminAccountProfileView,
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
  adminProfileSchema,
  changePasswordFormSchema,
  changePasswordSchema,
  customerProfileSchema,
  customerSelfProfileFormSchema,
  fullNamePhoneSelfProfileFormSchema,
  meSchema,
  staffProfileSchema,
  updateSelfProfileSchema,
  type ChangePasswordFormInput,
  type ChangePasswordInput,
  type CustomerSelfProfileFormValues,
  type FullNamePhoneSelfProfileFormValues,
  type UpdateSelfProfileInput,
} from "./schemas/index";
export type { AdminProfile, CustomerProfile, Me, StaffProfile } from "./types";
