/**
 * Auth feature — public API for app and cross-feature consumers.
 *
 * Internal: api/, components/, hooks/, provider/, schemas/
 * Client session infra: @/lib/auth (refresh manager — shared with browser-api-client)
 */
export { login, logout, register } from "./api";
export {
  AdminGate,
  BakerGate,
  CustomerGate,
  ManagerGate,
  PublicSessionGate,
  LoginForm,
  LogoutButton,
  MustChangePasswordGate,
  RegisterForm,
  StaffGate,
} from "./components";
export { useLogin, useLogout, useRegister } from "./hooks";
export type { LoginMutationVariables } from "./hooks";
export {
  loginSchema,
  registerSchema,
  tokenResponseSchema,
  userSchema,
} from "./schemas";
export type {
  LoginInput,
  RegisterInput,
  TokenResponse,
  User,
} from "./schemas";
