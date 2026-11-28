import { describe, expect, it } from "vitest";

import { ApiError, ApiErrorCode, apiErrorFromPayload, userFacingErrorMessage } from "./api-error";

function apiError(status: number, code: string): ApiError {
  return new ApiError(status, { code, message: code });
}

describe("ApiError.isAccessTokenRejected", () => {
  it("is true for the 401s the auth middleware gives an access token", () => {
    for (const code of [ApiErrorCode.Unauthorized, ApiErrorCode.TokenExpired, ApiErrorCode.SessionRevoked]) {
      expect(apiError(401, code).isAccessTokenRejected(), code).toBe(true);
    }
  });

  it("is false for a 401 about something else", () => {
    expect(apiError(401, ApiErrorCode.InvalidCredentials).isAccessTokenRejected()).toBe(false);
    expect(apiError(401, ApiErrorCode.Unknown).isAccessTokenRejected()).toBe(false);
  });

  it("is false outside a 401", () => {
    expect(apiError(403, ApiErrorCode.Unauthorized).isAccessTokenRejected()).toBe(false);
  });
});

describe("apiErrorFromPayload", () => {
  it("reads the error from an envelope", () => {
    const error = apiErrorFromPayload(422, {
      success: false,
      error: { code: ApiErrorCode.AccountHasOpenOrders, message: "Open orders" },
    });

    expect(error).toMatchObject({ status: 422, code: ApiErrorCode.AccountHasOpenOrders });
  });

  it("falls back to unknown when the body is not an envelope", () => {
    expect(apiErrorFromPayload(401, null)).toMatchObject({ status: 401, code: ApiErrorCode.Unknown });
  });
});

describe("userFacingErrorMessage", () => {
  const fallback = "We could not do that. Please try again.";

  it("gives the API's message for a refusal it explains", () => {
    const refusal = new ApiError(422, { code: ApiErrorCode.CartEmpty, message: "Cart is empty" });

    expect(userFacingErrorMessage(refusal, fallback)).toBe("Cart is empty");
  });

  it("gives one fixed line when the caller is rate limited", () => {
    expect(userFacingErrorMessage(apiError(429, ApiErrorCode.RateLimited), fallback)).toBe(
      "You are doing that too often. Wait a moment, then try again.",
    );
  });

  it("gives the fallback for a server failure, a timeout and an unreadable response", () => {
    expect(userFacingErrorMessage(apiError(500, ApiErrorCode.Internal), fallback)).toBe(fallback);
    expect(userFacingErrorMessage(apiError(504, ApiErrorCode.Timeout), fallback)).toBe(fallback);
    expect(userFacingErrorMessage(apiError(502, ApiErrorCode.InvalidResponse), fallback)).toBe(fallback);
  });

  it("gives the fallback for an error that is not the API's", () => {
    expect(userFacingErrorMessage(new TypeError("Failed to fetch"), fallback)).toBe(fallback);
  });
});
