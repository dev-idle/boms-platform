import { describe, expect, it } from "vitest";

import { loginSchema, registerSchema } from "./index";

const PASSWORD = "Choux-pastry-2026";

/**
 * People paste addresses with stray spaces; the form trims them before the
 * address is judged, and a bad address reads in the form's own words.
 */
describe.each([
  ["sign-in", loginSchema],
  ["registration", registerSchema],
])("%s email", (_, schema) => {
  it("accepts an address with surrounding spaces and submits it trimmed", () => {
    const result = schema.safeParse({ email: "  mai@example.com ", password: PASSWORD, accept_terms: true });
    expect(result.success).toBe(true);
    expect(result.success ? result.data.email : undefined).toBe("mai@example.com");
  });

  it("names a malformed address in the form's words", () => {
    const result = schema.safeParse({ email: "mai@", password: PASSWORD });
    expect(result.success).toBe(false);
    expect(result.success ? undefined : result.error.issues[0]?.message).toBe(
      "Enter a valid email address",
    );
  });
});

describe("registration consent", () => {
  it("refuses a sign-up that has not accepted the policies", () => {
    const result = registerSchema.safeParse({ email: "mai@example.com", password: PASSWORD, accept_terms: false });
    expect(result.success ? undefined : result.error.issues[0]?.message).toBe(
      "Accept the terms and the privacy policy to create an account",
    );
  });

  it("takes a sign-up that has", () => {
    expect(
      registerSchema.safeParse({ email: "mai@example.com", password: PASSWORD, accept_terms: true }).success,
    ).toBe(true);
  });
});

describe("registration email length", () => {
  it("names a malformed address before its length when it is both", () => {
    const result = registerSchema.safeParse({ email: `${"m".repeat(300)}@`, password: PASSWORD });
    expect(result.success ? undefined : result.error.issues[0]?.message).toBe(
      "Enter a valid email address",
    );
  });

  it("says an address is too long in words", () => {
    const result = registerSchema.safeParse({
      email: `mai@${"patisserie.".repeat(25)}com`,
      password: PASSWORD,
    });
    expect(result.success ? undefined : result.error.issues[0]?.message).toBe(
      "Email must be at most 255 characters",
    );
  });
});
