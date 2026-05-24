/**
 * Auth feature — public API for app and cross-feature consumers.
 *
 * Internal: api/, components/, hooks/, lib/validation-messages, provider/, schemas/
 * Client session infra: @/lib/auth (refresh manager — shared with browser-api-client)
 */
export { getMe, login, logout, register } from "./api";
export {
  AdminGate,
  AuthenticatedRedirect,
  LoginForm,
  LogoutButton,
  RegisterForm,
  RoleGate,
} from "./components";
export {
  authQueryKeys,
  meQueryOptions,
  useLogin,
  useLogout,
  useMe,
  useRegister,
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
