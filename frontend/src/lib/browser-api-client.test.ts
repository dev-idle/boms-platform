import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { ApiErrorCode } from "@/lib/errors";
import { useAuthStore } from "@/stores/auth-store";

import { browserRequest, browserRequestVoid } from "./browser-api-client";

const refreshNow = vi.hoisted(() => vi.fn());
vi.mock("@/lib/auth", () => ({ refreshNow }));

const OLD_TOKEN = "old-access-token";
const NEW_TOKEN = "new-access-token";

function envelope(status: number, body: object): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function refusal(code: string): Response {
  return envelope(401, { success: false, error: { code, message: code } });
}

const fetchMock = vi.fn<typeof fetch>();

/** The bearer token each request carried, in order. */
function sentTokens(): (string | null)[] {
  return fetchMock.mock.calls.map(([, init]) => new Headers(init?.headers).get("Authorization"));
}

beforeEach(() => {
  vi.stubGlobal("fetch", fetchMock);
  useAuthStore.setState({ accessToken: OLD_TOKEN });
  refreshNow.mockImplementation(async () => {
    useAuthStore.setState({ accessToken: NEW_TOKEN });
  });
});

afterEach(() => {
  vi.unstubAllGlobals();
  fetchMock.mockReset();
  refreshNow.mockReset();
});

describe("browser API client, on 401", () => {
  it.each([ApiErrorCode.TokenExpired, ApiErrorCode.Unauthorized, ApiErrorCode.SessionRevoked])(
    "refreshes once and retries when the access token is turned away (%s)",
    async (code) => {
      fetchMock
        .mockResolvedValueOnce(refusal(code))
        .mockResolvedValueOnce(envelope(200, { success: true, data: { ok: true } }));

      await expect(browserRequest("/api/v1/me")).resolves.toEqual({ ok: true });

      expect(refreshNow).toHaveBeenCalledOnce();
      expect(refreshNow).toHaveBeenCalledWith({ redirectOnFailure: false, staleToken: OLD_TOKEN });
      expect(sentTokens()).toEqual([`Bearer ${OLD_TOKEN}`, `Bearer ${NEW_TOKEN}`]);
    },
  );

  it("answers a wrong password once, without a refresh", async () => {
    fetchMock.mockResolvedValueOnce(refusal(ApiErrorCode.InvalidCredentials));

    await expect(
      browserRequestVoid("/api/v1/me/password", { method: "PATCH", json: { old_password: "guess" } }),
    ).rejects.toMatchObject({ status: 401, code: ApiErrorCode.InvalidCredentials });

    expect(fetchMock).toHaveBeenCalledOnce();
    expect(refreshNow).not.toHaveBeenCalled();
  });

  it("does not refresh for a 401 it cannot read", async () => {
    fetchMock.mockResolvedValueOnce(new Response("", { status: 401 }));

    await expect(browserRequest("/api/v1/me")).rejects.toMatchObject({
      status: 401,
      code: ApiErrorCode.Unknown,
    });

    expect(refreshNow).not.toHaveBeenCalled();
  });

  it("gives up after one retry", async () => {
    fetchMock
      .mockResolvedValueOnce(refusal(ApiErrorCode.TokenExpired))
      .mockResolvedValueOnce(refusal(ApiErrorCode.SessionRevoked));

    await expect(browserRequest("/api/v1/me")).rejects.toMatchObject({
      code: ApiErrorCode.SessionRevoked,
    });

    expect(refreshNow).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("leaves requests that opt out alone", async () => {
    fetchMock.mockResolvedValueOnce(refusal(ApiErrorCode.TokenExpired));

    await expect(
      browserRequest("/api/v1/catalog/products", { skipRefreshRetry: true }),
    ).rejects.toMatchObject({ code: ApiErrorCode.TokenExpired });

    expect(refreshNow).not.toHaveBeenCalled();
  });
});
