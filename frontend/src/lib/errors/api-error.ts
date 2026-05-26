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
  Conflict: "conflict",
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
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}
