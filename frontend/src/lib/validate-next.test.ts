import { describe, expect, it } from "vitest";

import { USER_ROLE } from "@/constants/roles";
import {
  loginHrefPreservingNext,
  validateNext,
  validateNextForRole,
} from "@/lib/validate-next";

describe("validateNext", () => {
  it("accepts safe relative paths", () => {
    expect(validateNext("/products")).toBe("/products");
    expect(validateNext("/admin/users?page=1")).toBe("/admin/users?page=1");
  });

  it("rejects open-redirect vectors", () => {
    expect(validateNext(null)).toBeNull();
    expect(validateNext("//evil.com")).toBeNull();
    expect(validateNext("https://evil.com")).toBeNull();
    expect(validateNext("/\\evil")).toBeNull();
    expect(validateNext("/%2f%2fevil.com")).toBeNull();
    expect(validateNext("/login@evil.com")).toBeNull();
    expect(validateNext("/path%00")).toBeNull();
  });
});

describe("loginHrefPreservingNext", () => {
  it("embeds a safe next path", () => {
    expect(loginHrefPreservingNext("/admin/users")).toBe(
      "/login?next=%2Fadmin%2Fusers",
    );
  });

  it("omits next for unsafe paths", () => {
    expect(loginHrefPreservingNext("//evil")).toBe("/login");
  });
});

describe("validateNextForRole", () => {
  it("allows paths in the role namespace only", () => {
    expect(validateNextForRole("/products", USER_ROLE.customer)).toBe(
      "/products",
    );
    expect(
      validateNextForRole("/staff/account/profile", USER_ROLE.staff),
    ).toBe("/staff/account/profile");
  });

  it("rejects cross-role paths", () => {
    expect(validateNextForRole("/admin", USER_ROLE.customer)).toBeNull();
    expect(validateNextForRole("/products", USER_ROLE.admin)).toBeNull();
  });
});
