import { describe, expect, it } from "vitest";

import { USER_ROLE } from "@/constants/roles";
import {
  loginHrefAfterRegister,
  loginHrefPreservingNext,
  loginHrefWithNext,
  registerHrefPreservingNext,
  registerHrefWithNext,
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

describe("registerHrefPreservingNext", () => {
  it("embeds a safe next path", () => {
    expect(
      registerHrefPreservingNext("/products", "?id=1"),
    ).toBe("/register?next=%2Fproducts%3Fid%3D1");
  });

  it("omits next for unsafe paths", () => {
    expect(registerHrefPreservingNext("//evil")).toBe("/register");
  });
});

describe("loginHrefAfterRegister", () => {
  it("includes registered flag and safe next", () => {
    expect(loginHrefAfterRegister("/cart")).toBe(
      "/login?registered=1&next=%2Fcart",
    );
  });

  it("rejects unsafe next", () => {
    expect(loginHrefAfterRegister("//evil")).toBe("/login?registered=1");
  });
});

describe("loginHrefWithNext", () => {
  it("returns login with validated next", () => {
    expect(loginHrefWithNext("/orders")).toBe("/login?next=%2Forders");
  });

  it("returns bare login for unsafe next", () => {
    expect(loginHrefWithNext("https://evil")).toBe("/login");
  });
});

describe("registerHrefWithNext", () => {
  it("returns register with validated next", () => {
    expect(registerHrefWithNext("/cart")).toBe("/register?next=%2Fcart");
  });

  it("returns bare register for unsafe next", () => {
    expect(registerHrefWithNext("//evil")).toBe("/register");
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
