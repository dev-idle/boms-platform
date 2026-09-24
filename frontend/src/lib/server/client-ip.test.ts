import { describe, expect, it } from "vitest";

import { resolveClientIp } from "./client-ip";

function headers(entries: Record<string, string>): Headers {
  return new Headers(entries);
}

describe("resolveClientIp", () => {
  it("prefers the single address the edge reports", () => {
    expect(
      resolveClientIp(
        headers({ "x-real-ip": "203.0.113.7", "x-forwarded-for": "198.51.100.9" }),
      ),
    ).toBe("203.0.113.7");
  });

  it("takes the client the chain starts with", () => {
    expect(
      resolveClientIp(headers({ "x-forwarded-for": "203.0.113.7, 10.0.0.1, 10.0.0.2" })),
    ).toBe("203.0.113.7");
  });

  it("reads IPv6", () => {
    expect(resolveClientIp(headers({ "x-real-ip": "2001:db8::1" }))).toBe("2001:db8::1");
  });

  it("returns nothing rather than a value the API would have to trust", () => {
    // The API falls back to the socket address when the header is absent, which
    // is the safe answer; a forged or malformed value must never reach it.
    for (const value of ["", "   ", "not-an-ip", "999.1.1.1", "1.2.3.4 OR 1=1", "<script>"]) {
      expect(resolveClientIp(headers({ "x-real-ip": value })), value).toBeUndefined();
    }
    expect(resolveClientIp(headers({}))).toBeUndefined();
  });

  it("falls through to the chain when the single header is junk", () => {
    expect(
      resolveClientIp(
        headers({ "x-real-ip": "not-an-ip", "x-forwarded-for": "203.0.113.7" }),
      ),
    ).toBe("203.0.113.7");
  });
});
