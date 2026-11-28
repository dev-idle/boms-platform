/**
 * Auth feature — public API for app and cross-feature consumers.
 *
 * Internal: api/, components/, hooks/, provider/, schemas/
 * Client session infra: @/lib/auth (refresh manager — shared with browser-api-client)
 */
export {
  AdminGate,
  AuthFormShell,
  AuthFormSkeleton,
  AuthLayoutFrame,
  BakerGate,
  CustomerGate,
  ForgotPasswordForm,
  LoginForm,
  ManagerGate,
  PublicSessionGate,
  RegisterForm,
  ResetPasswordForm,
  SignOutDialog,
  StaffGate,
  UnsubscribeView,
  VerifyEmailView,
} from "./components";
export { AUTH_FORM_COPY } from "./lib/auth-form-copy";
