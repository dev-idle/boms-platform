import { describe, expect, it } from "vitest";

import { USER_ROLE } from "@/constants/roles";

import { createOperationalSchema, updateRoleSchema } from "./index";

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

  it("accepts a named role change, phone still optional", () => {
    const result = updateRoleSchema.safeParse({
      role: USER_ROLE.manager,
      full_name: "Mai Tran",
      phone: "",
    });
    expect(result.success).toBe(true);
  });
});
