/**
 * User feature — self-service profile, password, and account lifecycle.
 *
 * Internal: api/, components/, hooks/, schemas/, types/
 * Session identity: re-exported via auth; `/api/v1/me` is the source of truth.
 */
export { getMe } from "./api";
export { primeMeQueryCache } from "./lib/prime-me-cache";
export {
  AdminAccountProfileView,
  CustomerAccountView,
  EmailVerificationNotice,
  OperationalAccountProfileView,
} from "./components";
export {
  meQueryOptions,
  userQueryKeys,
} from "./hooks";
