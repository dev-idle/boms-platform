/**
 * Auth feature — public API for app and cross-feature consumers.
 *
 * Internal: api/, components/, hooks/, provider/, schemas/
 * Client session infra: @/lib/auth (refresh manager — shared with browser-api-client)
 */
export { login, logout, register } from "./api";
export { getMe } from "@/features/user/api";
export {
  AdminGate,
  BakerGate,
  CustomerGate,
  ManagerGate,
  AuthenticatedRedirect,
  LoginForm,
  LogoutButton,
  MustChangePasswordGate,
  RegisterForm,
  StaffGate,
} from "./components";
export {
  meQueryOptions,
  useLogin,
  useLogout,
  useMe,
  useRegister,
  userQueryKeys,
} from "./hooks";
export type { LoginMutationVariables } from "./hooks";
export { AuthBootstrap, AuthProvider, useAuthStatus } from "./provider";
export {
  loginSchema,
  refreshResponseSchema,
  registerSchema,
  tokenResponseSchema,
  userSchema,
} from "./schemas";
export type {
  LoginInput,
  RefreshResponse,
  RegisterInput,
  TokenResponse,
  User,
} from "./schemas";
