import { describe, expect, it } from "vitest";

import { USER_ROLE } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";

import {
  homeRouteForRole,
  isPathAllowedForRole,
  isProtectedPath,
  passwordRouteForRole,
} from "./role-routes";

describe("homeRouteForRole", () => {
  it("returns the canonical home per role", () => {
    expect(homeRouteForRole(USER_ROLE.customer)).toBe(ROUTE.products);
    expect(homeRouteForRole(USER_ROLE.admin)).toBe(ROUTE.admin.dashboard);
    expect(homeRouteForRole(USER_ROLE.staff)).toBe(
      ROUTE.staff.account.profile,
    );
  });
});

describe("isProtectedPath", () => {
  it("includes all role namespaces and excludes public auth pages", () => {
    expect(isProtectedPath("/products")).toBe(true);
    expect(isProtectedPath("/admin/users")).toBe(true);
    expect(isProtectedPath("/login")).toBe(false);
    expect(isProtectedPath("/")).toBe(false);
  });
});

describe("isPathAllowedForRole", () => {
  it("matches prefix boundaries", () => {
    expect(isPathAllowedForRole("/baker/account/profile", USER_ROLE.baker)).toBe(
      true,
    );
    expect(isPathAllowedForRole("/baker-evil", USER_ROLE.baker)).toBe(false);
    expect(isPathAllowedForRole("/admin/users", USER_ROLE.customer)).toBe(
      false,
    );
  });

  it("denies every other role namespace (including admin)", () => {
    const cases = [
      { path: "/admin", role: USER_ROLE.customer },
      { path: "/products", role: USER_ROLE.admin },
      { path: "/staff/account/profile", role: USER_ROLE.baker },
      { path: "/baker/account/profile", role: USER_ROLE.manager },
      { path: "/manager/account/profile", role: USER_ROLE.staff },
      { path: "/orders", role: USER_ROLE.staff },
    ] as const;

    for (const { path, role } of cases) {
      expect(isPathAllowedForRole(path, role)).toBe(false);
    }
  });

  it("allows only own namespace per role", () => {
    expect(isPathAllowedForRole("/admin/users", USER_ROLE.admin)).toBe(true);
    expect(isPathAllowedForRole("/cart", USER_ROLE.customer)).toBe(true);
    expect(isPathAllowedForRole("/staff/account/password", USER_ROLE.staff)).toBe(
      true,
    );
  });
});

describe("passwordRouteForRole", () => {
  it("returns role-scoped password routes", () => {
    expect(passwordRouteForRole(USER_ROLE.manager)).toBe(
      ROUTE.manager.account.password,
    );
    expect(passwordRouteForRole(USER_ROLE.admin)).toBe(
      ROUTE.admin.account.profile,
    );
  });
});
