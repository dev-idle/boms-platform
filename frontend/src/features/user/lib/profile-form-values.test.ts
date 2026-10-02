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
      customerProfileSnapshot({ display_name: "  Pat  ", phone: null, marketing_consent_at: null }),
    ).toEqual({
      display_name: "Pat",
      phone: "",
      marketing_opt_in: false,
    });
    expect(
      customerProfileSnapshot({ display_name: null, phone: null, marketing_consent_at: "2026-10-02T09:00:00+07:00" })
        .marketing_opt_in,
    ).toBe(true);
  });
});
