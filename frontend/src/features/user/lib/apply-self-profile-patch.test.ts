import { describe, expect, it } from "vitest";

import { USER_ROLE } from "@/constants/roles";

import { applySelfProfilePatch } from "./apply-self-profile-patch";
import type { Me } from "../types";

const baseMe = {
  id: "00000000-0000-4000-8000-000000000001",
  email: "user@example.com",
  email_verified: true,
  must_change_password: false,
  disabled: false,
  created_at: "2026-01-01T00:00:00.000Z",
} as const;

describe("applySelfProfilePatch", () => {
  it("updates admin full_name and phone", () => {
    const current: Me = {
      ...baseMe,
      role: USER_ROLE.admin,
      profile: {
        type: "admin",
        full_name: "Before",
        phone: null,
      },
    };

    const next = applySelfProfilePatch(current, {
      full_name: "After",
      phone: "+1",
    });

    expect(next.role).toBe(USER_ROLE.admin);
    if (next.role !== USER_ROLE.admin) {
      return;
    }
    expect(next.profile.full_name).toBe("After");
    expect(next.profile.phone).toBe("+1");
  });

  it("leaves omitted fields unchanged for operational roles", () => {
    const current: Me = {
      ...baseMe,
      role: USER_ROLE.staff,
      profile: {
        type: "staff",
        full_name: "Jane",
        phone: "555",
        employee_code: "E1",
        hire_date: "2024-01-01",
      },
    };

    const next = applySelfProfilePatch(current, { phone: null });

    expect(next.role).toBe(USER_ROLE.staff);
    if (next.role !== USER_ROLE.staff) {
      return;
    }
    expect(next.profile.full_name).toBe("Jane");
    expect(next.profile.phone).toBeNull();
    expect(next.profile.employee_code).toBe("E1");
  });
});
