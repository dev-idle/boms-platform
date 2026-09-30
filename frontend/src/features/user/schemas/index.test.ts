import { describe, expect, it } from "vitest";

import { dataExportSchema } from "./index";

const exported = {
  exported_at: "2026-09-29T08:15:00Z",
  account: {
    id: "00000000-0000-4000-8000-000000000001",
    email: "mai@example.com",
    role: "customer",
    profile: { type: "customer", display_name: "Mai", phone: "+84901234567" },
  },
  terms_acceptance: { version: "2026-09-29", accepted_at: "2026-09-29T08:00:00Z" },
  sessions: [{ signed_in_at: "2026-09-29T08:00:00Z", ip: "203.0.113.7", user_agent: "Test browser" }],
  account_activity: [{ action: "me.updated_profile", at: "2026-09-29T08:05:00Z", by: "you", before: {}, after: {} }],
  cart: [{ id: "00000000-0000-4000-8000-000000000003", line_type: "product", quantity: 1, added_at: "2026-09-29T08:10:00Z" }],
  orders: [
    {
      id: "00000000-0000-4000-8000-000000000002",
      code: "CH-260929-001",
      terms_acceptance: { version: "2026-09-29", accepted_at: "2026-09-29T08:01:00Z" },
      items: [{ name: "Croissant", quantity: 2 }],
      timeline: [{ status: "pending", at: "2026-09-29T08:01:00Z" }],
    },
  ],
};

describe("dataExportSchema", () => {
  it("hands over every field the API sends, however deep", () => {
    expect(dataExportSchema.parse(exported)).toEqual(exported);
  });

  it("refuses a response that is not an export", () => {
    expect(dataExportSchema.safeParse({ account: exported.account }).success).toBe(false);
  });
});
