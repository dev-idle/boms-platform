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
    // Stored numbers are E.164; the form holds the national number after +84.
    expect(
      fullNamePhoneSnapshotFromProfile("  Admin ", "+84912345678"),
    ).toEqual({
      full_name: "Admin",
      phone: "912345678",
    });
    expect(
      customerProfileSnapshot({ display_name: "  Pat  ", phone: null }),
    ).toEqual({
      display_name: "Pat",
      phone: "",
    });
  });
});
