import { describe, expect, it } from "vitest";

import { parseApiEnvelope } from "./api-envelope";

describe("parseApiEnvelope", () => {
  it("keeps meta fields it does not know, such as pagination", () => {
    const result = parseApiEnvelope({
      success: true,
      data: [],
      meta: { request_id: "req-1", pagination: { page: 1, page_size: 20, total: 0 } },
    });
    expect(result.success).toBe(true);
    expect(result.success ? result.data.meta : undefined).toMatchObject({
      pagination: { page: 1, page_size: 20, total: 0 },
    });
  });

  it("reads field details as text keyed by field", () => {
    const result = parseApiEnvelope({
      success: false,
      error: { code: "validation_error", message: "Invalid", details: { phone: "phone_format" } },
    });
    expect(result.success ? result.data.error?.details : undefined).toEqual({
      phone: "phone_format",
    });
  });

  it("rejects details that are not text", () => {
    const result = parseApiEnvelope({
      success: false,
      error: { code: "validation_error", message: "Invalid", details: { phone: 1 } },
    });
    expect(result.success).toBe(false);
  });
});
