import type { ApiErrorBody } from "@/lib/api-envelope";

/** Stable error codes returned by the backend (snake_case). Keep in sync with `internal/shared/errors`. */
export const ApiErrorCode = {
  Unauthorized: "unauthorized",
  TokenExpired: "token_expired",
  InvalidCredentials: "invalid_credentials",
  InvalidRefreshToken: "invalid_refresh_token",
  MissingRefreshToken: "missing_refresh_token",
  SessionRevoked: "session_revoked",
  Forbidden: "forbidden",
  PasswordChangeRequired: "password_change_required",
  NotFound: "not_found",
  ProfileNotFound: "profile_not_found",
  MeNotFound: "me_not_found",
  Conflict: "conflict",
  EmailExists: "email_exists",
  CannotModifySelf: "cannot_modify_self",
  InvalidRoleTransition: "invalid_role_transition",
  EmployeeCodeExists: "employee_code_exists",
  CategoryHasProducts: "category_has_products",
  SlugExists: "slug_exists",
  CodeExists: "code_exists",
  DiscountInactive: "discount_inactive",
  DiscountExpired: "discount_expired",
  DiscountExhausted: "discount_exhausted",
  DiscountMinOrderNotMet: "discount_min_order_not_met",
  CartEmpty: "cart_empty",
  CartMaxItems: "cart_max_items",
  ProductUnavailable: "product_unavailable",
  ComboUnavailable: "combo_unavailable",
  InvalidOrderStatusTransition: "invalid_order_status_transition",
  Validation: "validation_error",
  RateLimited: "rate_limited",
  Internal: "internal_error",
  // Client-synthesised codes (never returned by the backend).
  InvalidResponse: "invalid_response",
  Timeout: "timeout",
  Unknown: "unknown_error",
} as const;

const SESSION_ERROR_CODES = new Set<string>([
  ApiErrorCode.InvalidRefreshToken,
  ApiErrorCode.SessionRevoked,
  ApiErrorCode.MissingRefreshToken,
]);

export class ApiError extends Error {
  readonly code: string;
  readonly status: number;
  readonly details?: Record<string, string>;

  constructor(status: number, body: ApiErrorBody) {
    super(body.message);
    this.name = "ApiError";
    this.code = body.code;
    this.status = status;
    this.details = body.details;
  }

  /** True for any 401 response or a refresh-token-related code. */
  isAuthError(): boolean {
    return this.status === 401 || SESSION_ERROR_CODES.has(this.code);
  }

  /** True only when the failure means "access token expired" (FE should refresh). */
  isTokenExpired(): boolean {
    return this.code === ApiErrorCode.TokenExpired;
  }

  /** True when the user must change their password before the request can succeed. */
  isPasswordChangeRequired(): boolean {
    return this.code === ApiErrorCode.PasswordChangeRequired;
  }

  /** True for 403 Forbidden (insufficient role / permission). */
  isForbidden(): boolean {
    return this.status === 403 && this.code === ApiErrorCode.Forbidden;
  }

  /** True for backend validation failures (400 validation_error). */
  isValidation(): boolean {
    return this.status === 400 && this.code === ApiErrorCode.Validation;
  }

  /** True when the API returned per-field validation details. */
  hasValidationDetails(): boolean {
    return (
      this.details !== undefined &&
      Object.keys(this.details).length > 0 &&
      (this.isValidation() || this.status === 422)
    );
  }

  isCannotModifySelf(): boolean {
    return this.code === ApiErrorCode.CannotModifySelf;
  }

  isInvalidRoleTransition(): boolean {
    return this.code === ApiErrorCode.InvalidRoleTransition;
  }

  isEmployeeCodeExists(): boolean {
    return this.code === ApiErrorCode.EmployeeCodeExists;
  }

  isInvalidCredentials(): boolean {
    return this.code === ApiErrorCode.InvalidCredentials;
  }

  isEmailExists(): boolean {
    return this.code === ApiErrorCode.EmailExists;
  }

  isSlugExists(): boolean {
    return this.code === ApiErrorCode.SlugExists;
  }

  isCodeExists(): boolean {
    return this.code === ApiErrorCode.CodeExists;
  }
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}
