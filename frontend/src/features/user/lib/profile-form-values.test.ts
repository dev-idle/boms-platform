import { describe, expect, it } from "vitest";

import {
  customerProfileSnapshot,
  fullNamePhoneFormValuesEqual,
  fullNamePhoneSnapshotFromProfile,
  normalizeFullNamePhoneFormValues,
} from "./profile-form-values";

describe("profile-form-values", () => {
  it("normalizes trim and empty phone", () => {
    expect(
      normalizeFullNamePhoneFormValues({
        full_name: "  Admin  ",
        phone: "  ",
      }),
    ).toEqual({
      full_name: "Admin",
      phone: "",
    });
  });

  it("detects unsaved phone changes after normalization", () => {
    const baseline = { full_name: "Admin", phone: "555-0100" };
    expect(
      fullNamePhoneFormValuesEqual(
        { full_name: "Admin", phone: "555-0199" },
        baseline,
      ),
    ).toBe(false);
    expect(
      fullNamePhoneFormValuesEqual(
        { full_name: "Admin", phone: " 555-0100 " },
        baseline,
      ),
    ).toBe(true);
  });

  it("builds normalized snapshots from profile fields", () => {
    expect(
      fullNamePhoneSnapshotFromProfile("  Admin ", " 555-0100 "),
    ).toEqual({
      full_name: "Admin",
      phone: "555-0100",
    });
    expect(
      customerProfileSnapshot({ display_name: "  Pat  ", phone: null }),
    ).toEqual({
      display_name: "Pat",
      phone: "",
    });
  });
});
