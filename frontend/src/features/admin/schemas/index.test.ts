import { describe, expect, it } from "vitest";

import { USER_ROLE } from "@/constants/roles";

import {
  closedDateFormSchema,
  createOperationalSchema,
  storeSettingsFormSchema,
  updateRoleSchema,
} from "./index";

/**
 * An operational full name is the identity shown in the sidebar and in the
 * admin table. Create has always required one; changing a role used to accept
 * an empty name that the API then dropped without saying so.
 */
describe("operational full name", () => {
  it("is required when an admin changes a role", () => {
    const result = updateRoleSchema.safeParse({
      role: USER_ROLE.staff,
      full_name: "  ",
      phone: "",
    });
    expect(result.success).toBe(false);
    expect(result.success ? [] : result.error.issues.map((i) => i.message)).toContain(
      "Full name is required",
    );
  });

  it("is required when an admin creates a user", () => {
    const result = createOperationalSchema.safeParse({
      email: "mai@example.com",
      role: USER_ROLE.staff,
      full_name: "",
      phone: "",
    });
    expect(result.success).toBe(false);
    expect(result.success ? [] : result.error.issues.map((i) => i.message)).toContain(
      "Full name is required",
    );
  });

  it("names a malformed email before its length when it is both", () => {
    const result = createOperationalSchema.safeParse({
      email: `${"m".repeat(300)}@`,
      role: USER_ROLE.staff,
      full_name: "Mai Tran",
      phone: "",
    });
    expect(result.success).toBe(false);
    expect(result.success ? undefined : result.error.issues[0]?.message).toBe(
      "Enter a valid email",
    );
  });

  it("accepts a named role change, phone still optional", () => {
    const result = updateRoleSchema.safeParse({
      role: USER_ROLE.manager,
      full_name: "Mai Tran",
      phone: "",
    });
    expect(result.success).toBe(true);
  });
});

describe("store settings form", () => {
  const valid = {
    opens_at: "08:00",
    closes_at: "18:00",
    preorder_min_lead_minutes: 120,
    max_advance_days: 14,
    slot_minutes: 30,
    slot_capacity: 10,
    instant_prep_minutes: 20,
  };

  it("accepts the seeded rules", () => {
    expect(storeSettingsFormSchema.safeParse(valid).success).toBe(true);
  });

  it("names the field an impossible rule breaks", () => {
    const paths = (input: object) => {
      const result = storeSettingsFormSchema.safeParse({ ...valid, ...input });
      return result.success ? [] : result.error.issues.map((issue) => issue.path.join("."));
    };
    expect(paths({ closes_at: "07:00" })).toEqual(["closes_at"]);
    expect(paths({ opens_at: "17:45" })).toEqual(["slot_minutes"]);
    expect(paths({ slot_minutes: 45 })).toEqual(["slot_minutes"]);
    expect(paths({ slot_capacity: 0 })).toEqual(["slot_capacity"]);
    expect(paths({ instant_prep_minutes: 241 })).toEqual(["instant_prep_minutes"]);
    expect(paths({ max_advance_days: 1, preorder_min_lead_minutes: 1440 })).toEqual([
      "preorder_min_lead_minutes",
    ]);
    expect(paths({ max_advance_days: 91 })).toEqual(["max_advance_days"]);
  });
});

describe("closed day form", () => {
  it("accepts a day with a plain reason", () => {
    expect(closedDateFormSchema.safeParse({ date: "2026-10-20", reason: "Staff training" }).success).toBe(true);
  });

  it("refuses control and invisible formatting characters", () => {
    for (const reason of ["Closed\nearly", "Tet \u202eyadiloh", "Zero\u200bwidth"]) {
      expect(closedDateFormSchema.safeParse({ date: "2026-10-20", reason }).success).toBe(false);
    }
  });
});
