/**
 * Auth feature — public API for app and cross-feature consumers.
 *
 * Internal: api/, components/, hooks/, provider/, schemas/
 * Client session infra: @/lib/auth (refresh manager — shared with browser-api-client)
 */
export {
  AdminGate,
  BakerGate,
  CustomerGate,
  LoginForm,
  LogoutButton,
  ManagerGate,
  PublicSessionGate,
  RegisterForm,
  StaffGate,
} from "./components";
