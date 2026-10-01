import type { ApiEnvelope } from "@/lib/api-envelope";
import { parseApiEnvelope, type ApiErrorBody } from "@/lib/api-envelope";

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
  UserNotFound: "user_not_found",
  ServiceUnavailable: "service_unavailable",
  Conflict: "conflict",
  EmailExists: "email_exists",
  CannotModifySelf: "cannot_modify_self",
  CannotModifyAdmin: "cannot_modify_admin",
  InvalidRoleTransition: "invalid_role_transition",
  EmployeeCodeExists: "employee_code_exists",
  PhoneExists: "phone_exists",
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
  PickupTooSoon: "pickup_too_soon",
  PickupTooFar: "pickup_too_far",
  PickupClosedDay: "pickup_closed_day",
  PickupSoldOut: "pickup_sold_out",
  PickupOutsideHours: "pickup_outside_hours",
  PickupOffSlot: "pickup_off_slot",
  PickupSlotFull: "pickup_slot_full",
  PickupDayLimit: "pickup_day_limit",
  TermsNotAccepted: "terms_not_accepted",
  AccountHasOpenOrders: "account_has_open_orders",
  AccountErased: "account_erased",
  EmailNotVerified: "email_not_verified",
  EmailAlreadyVerified: "email_already_verified",
  InvalidLink: "invalid_link",
  DiscountUsedUp: "discount_used_up",
  OrderNotPayable: "order_not_payable",
  PaymentNotCompleted: "payment_not_completed",
  PaymentUnderReview: "payment_under_review",
  PickupCodeInvalid: "pickup_code_invalid",
  PickupCodeLocked: "pickup_code_locked",
  WebhookInvalid: "webhook_invalid",
  InvalidTicketTransition: "invalid_ticket_transition",
  TicketOrderNotActive: "ticket_order_not_active",
  TicketNotMovable: "ticket_not_movable",
  TicketStationTaken: "ticket_station_taken",
  ClosedDateExists: "closed_date_exists",
  Validation: "validation_error",
  RateLimited: "rate_limited",
  Internal: "internal_error",
  // Client-synthesised codes (never returned by the backend).
  InvalidResponse: "invalid_response",
  Timeout: "timeout",
  Unknown: "unknown_error",
} as const;

/**
 * The 401 codes the auth middleware answers when it turns the access token away
 * (`backend/internal/middleware/auth.go`): missing or invalid, expired, or its
 * session rotated or ended. A refresh can cure these; no other 401 is about the
 * token — `invalid_credentials` is a wrong password, and the answer itself.
 */
const ACCESS_TOKEN_REJECTION_CODES = new Set<string>([
  ApiErrorCode.Unauthorized,
  ApiErrorCode.TokenExpired,
  ApiErrorCode.SessionRevoked,
]);

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

  /** True when the API turned the access token away, so a refresh may cure it. */
  isAccessTokenRejected(): boolean {
    return this.status === 401 && ACCESS_TOKEN_REJECTION_CODES.has(this.code);
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

  isCannotModifyAdmin(): boolean {
    return this.code === ApiErrorCode.CannotModifyAdmin;
  }

  isInvalidRoleTransition(): boolean {
    return this.code === ApiErrorCode.InvalidRoleTransition;
  }

  isEmployeeCodeExists(): boolean {
    return this.code === ApiErrorCode.EmployeeCodeExists;
  }

  isPhoneExists(): boolean {
    return this.code === ApiErrorCode.PhoneExists;
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

  /** True for HTTP 429 / rate_limited (transient — do not treat as logout). */
  isRateLimited(): boolean {
    return this.status === 429 || this.code === ApiErrorCode.RateLimited;
  }
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}

/** Maps a parsed HTTP failure envelope to ApiError (browser + refresh paths). */
export function throwApiErrorFromEnvelope(
  status: number,
  envelope: ApiEnvelope,
): never {
  if (envelope.error) {
    throw new ApiError(status, envelope.error);
  }
  throw new ApiError(status, {
    code: ApiErrorCode.Unknown,
    message: `Request failed with HTTP ${status}`,
  });
}

/** Maps an unvalidated JSON payload to ApiError (browser fetch path). */
export function apiErrorFromPayload(status: number, payload: unknown): ApiError {
  const envelope = parseApiEnvelope(payload);
  if (envelope.success && envelope.data.error) {
    return new ApiError(status, envelope.data.error);
  }
  return new ApiError(status, {
    code: ApiErrorCode.Unknown,
    message: `Request failed with HTTP ${status}`,
  });
}
