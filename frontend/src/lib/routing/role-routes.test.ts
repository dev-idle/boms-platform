import { describe, expect, it } from "vitest";

import { USER_ROLE } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";

import {
  homeRouteForRole,
  isPathAllowedForRole,
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
});

describe("passwordRouteForRole", () => {
  it("returns role-scoped password routes", () => {
    expect(passwordRouteForRole(USER_ROLE.manager)).toBe(
      ROUTE.manager.account.password,
    );
  });
});
