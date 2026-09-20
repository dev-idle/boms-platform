import { describe, expect, it } from "vitest";

import { USER_ROLE } from "@/constants/roles";

import {
  adminUserAccountStatus,
  adminUserRowActions,
  isAdminAccountManagementLocked,
} from "./account-management";

describe("isAdminAccountManagementLocked", () => {
  it("locks admin accounts only", () => {
    expect(isAdminAccountManagementLocked({ role: USER_ROLE.admin })).toBe(true);
    expect(isAdminAccountManagementLocked({ role: USER_ROLE.manager })).toBe(false);
    expect(isAdminAccountManagementLocked({ role: USER_ROLE.customer })).toBe(false);
  });
});

describe("adminUserAccountStatus", () => {
  it("reports admin rows as protected regardless of the disabled flag", () => {
    expect(adminUserAccountStatus({ disabled: false, role: USER_ROLE.admin })).toEqual({
      label: "Protected",
      variant: "completed",
    });
    expect(adminUserAccountStatus({ disabled: true, role: USER_ROLE.admin })).toEqual({
      label: "Protected",
      variant: "completed",
    });
  });

  it("distinguishes active from disabled for mutable roles", () => {
    expect(adminUserAccountStatus({ disabled: false, role: USER_ROLE.manager })).toEqual({
      label: "Active",
      variant: "ready",
    });
    expect(adminUserAccountStatus({ disabled: true, role: USER_ROLE.customer })).toEqual({
      label: "Disabled",
      variant: "cancelled",
    });
  });
});

describe("adminUserRowActions", () => {
  const staff = { disabled: false, id: "u-staff", role: USER_ROLE.staff };

  it("keeps both slots live for an active mutable account", () => {
    expect(adminUserRowActions(staff, "u-admin")).toEqual({
      account: { kind: "live", label: "Disable" },
      sessions: { kind: "live" },
    });
  });

  it("offers Enable and greys Revoke for a disabled account", () => {
    const result = adminUserRowActions({ ...staff, disabled: true }, "u-admin");
    expect(result.account).toEqual({ kind: "live", label: "Enable" });
    expect(result.sessions.kind).toBe("blocked");
  });

  it("greys both slots on the signed-in row before the admin-role rule", () => {
    const self = { disabled: false, id: "u-admin", role: USER_ROLE.admin };
    const result = adminUserRowActions(self, "u-admin");
    expect(result.account).toMatchObject({
      kind: "blocked",
      label: "Disable",
      reason: "You cannot disable your own account.",
    });
    expect(result.sessions).toMatchObject({
      kind: "blocked",
      reason: "You cannot revoke your own sessions.",
    });
  });

  it("greys both slots for another admin account", () => {
    const other = { disabled: false, id: "u-other", role: USER_ROLE.admin };
    const result = adminUserRowActions(other, "u-admin");
    expect(result.account.kind).toBe("blocked");
    expect(result.sessions.kind).toBe("blocked");
  });
});
