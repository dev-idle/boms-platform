import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("server-only", () => ({}));
vi.mock("next/headers", () => ({
  cookies: async () => ({ getAll: () => [] }),
}));
vi.mock("next/server", () => ({ connection: async () => undefined }));
vi.mock("@/lib/env", () => ({
  getServerEnv: () => ({
    BOMS_BACKEND_URL: "http://backend.test",
    INTERNAL_PROXY_SECRET: "test-internal-secret",
  }),
}));

import { proxyRequestToBackend } from "./backend-proxy";

/** Headers the upstream Fiber API received for one browser request. */
async function forwardedHeaders(browserHeaders: HeadersInit): Promise<Headers> {
  const fetchSpy = vi
    .spyOn(globalThis, "fetch")
    .mockResolvedValue(new Response("{}", { status: 200 }));

  await proxyRequestToBackend(
    new Request("http://app.test/api/v1/catalog/products", {
      headers: browserHeaders,
    }),
    ["catalog", "products"],
  );

  const [, init] = fetchSpy.mock.calls[0] as [string, RequestInit];
  return new Headers(init.headers);
}

afterEach(() => {
  vi.restoreAllMocks();
});

describe("proxyRequestToBackend", () => {
  it("forwards the address the edge proxy stamped", async () => {
    const headers = await forwardedHeaders({ "x-client-ip": "203.0.113.7" });

    expect(headers.get("x-client-ip")).toBe("203.0.113.7");
  });

  it("never reads or forwards the raw forwarding headers", async () => {
    // proxy.ts resolved them into X-Client-IP already; reading them again here
    // would take a value the edge proxy dropped.
    const headers = await forwardedHeaders({
      "x-forwarded-for": "9.9.9.9",
      "x-real-ip": "9.9.9.9",
      forwarded: "for=9.9.9.9",
    });

    expect(headers.get("x-client-ip")).toBeNull();
    expect(headers.get("x-forwarded-for")).toBeNull();
    expect(headers.get("x-real-ip")).toBeNull();
    expect(headers.get("forwarded")).toBeNull();
  });

  it("sends no address when the stamp is not one", async () => {
    const headers = await forwardedHeaders({ "x-client-ip": "1.2.3.4 OR 1=1" });

    expect(headers.get("x-client-ip")).toBeNull();
  });

  it("still injects the internal secret and a request id", async () => {
    const headers = await forwardedHeaders({ "x-internal-secret": "forged" });

    expect(headers.get("x-internal-secret")).toBe("test-internal-secret");
    expect(headers.get("x-request-id")).toBeTruthy();
  });
});
