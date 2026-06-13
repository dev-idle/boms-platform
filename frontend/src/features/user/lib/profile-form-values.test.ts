import { describe, expect, it } from "vitest";

import {
  customerProfileSnapshot,
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
