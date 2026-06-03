import { describe, expect, it } from "vitest";

import { USER_ROLE } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";

import { resolvePostAuthDestination } from "./post-auth-destination";

describe("resolvePostAuthDestination", () => {
  it("returns role home when no next path", () => {
    expect(resolvePostAuthDestination(USER_ROLE.admin)).toBe(
      ROUTE.admin.dashboard,
    );
    expect(resolvePostAuthDestination(USER_ROLE.customer)).toBe(ROUTE.products);
  });

  it("honors next when allowed for the role", () => {
    expect(
      resolvePostAuthDestination(USER_ROLE.admin, { next: "/admin/users" }),
    ).toBe("/admin/users");
  });

  it("rejects next outside the role namespace", () => {
    expect(
      resolvePostAuthDestination(USER_ROLE.baker, { next: "/admin" }),
    ).toBe(ROUTE.baker.account.profile);
  });

  it("prioritizes forced password change over next and home", () => {
    expect(
      resolvePostAuthDestination(USER_ROLE.staff, {
        next: "/staff/account/profile",
        mustChangePassword: true,
      }),
    ).toBe(ROUTE.staff.account.password);
    expect(
      resolvePostAuthDestination(USER_ROLE.admin, {
        next: "/admin/users",
        mustChangePassword: true,
      }),
    ).toBe(ROUTE.admin.account.profile);
  });
});
