import { NextRequest } from "next/server";
import type { NextFetchEvent } from "next/server";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/lib/env", () => ({
  getServerEnv: () => ({
    BOMS_BACKEND_URL: "http://backend.test",
    INTERNAL_PROXY_SECRET: "test-internal-secret",
  }),
  getBackendOrigin: () => "http://backend.test",
}));

import { config, proxy } from "./proxy";

/** The request headers proxy.ts hands on to the route handler for one call. */
async function forwarded(headers: Record<string, string>, path = "/api/v1/cart") {
  const response = await proxy(new NextRequest(`http://app.test${path}`, { headers }), {} as NextFetchEvent);
  if (!response) {
    throw new Error("proxy returned no response");
  }
  return (name: string) => response.headers.get(`x-middleware-request-${name}`);
}

// The API keys rate limits and audit rows by the visitor's address. Only the
// edge knows it, and only this proxy may turn that report into X-Client-IP.
describe("proxy client address", () => {
  it("stamps the address the edge reported", async () => {
    const header = await forwarded({ "x-real-ip": "203.0.113.7" });
    expect(header("x-client-ip")).toBe("203.0.113.7");
  });

  it("replaces an address the browser claims for itself", async () => {
    const header = await forwarded({ "x-client-ip": "9.9.9.9", "x-forwarded-for": "203.0.113.7" });
    expect(header("x-client-ip")).toBe("203.0.113.7");
  });

  it("stamps nothing when the edge reported nothing usable", async () => {
    const header = await forwarded({ "x-client-ip": "9.9.9.9", "x-real-ip": "not-an-ip" });
    expect(header("x-client-ip")).toBeNull();
  });

  it("hands on none of the raw forwarding headers", async () => {
    const header = await forwarded({ "x-real-ip": "203.0.113.7", forwarded: "for=9.9.9.9" });
    expect(header("x-real-ip")).toBeNull();
    expect(header("forwarded")).toBeNull();
  });

  it("runs on every API route, so no request reaches the BFF unstamped", () => {
    const matcher = new RegExp(`^${config.matcher[0]}$`);
    expect(matcher.test("/api/v1/cart")).toBe(true);
    expect(matcher.test("/api/v1/auth/login")).toBe(true);
    expect(matcher.test("/_next/static/chunk.js")).toBe(false);
  });
});
