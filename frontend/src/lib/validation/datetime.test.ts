import { describe, expect, it } from "vitest";

import { apiDateTimeSchema } from "./datetime";

describe("apiDateTimeSchema", () => {
  it("accepts UTC Z suffix", () => {
    expect(
      apiDateTimeSchema.safeParse("2026-06-17T09:50:00.000Z").success,
    ).toBe(true);
  });

  it("accepts numeric timezone offset from Go JSON", () => {
    expect(
      apiDateTimeSchema.safeParse("2026-06-17T09:50:00+07:00").success,
    ).toBe(true);
  });

  it("rejects datetime-local without offset", () => {
    expect(apiDateTimeSchema.safeParse("2026-06-17T09:50:00").success).toBe(
      false,
    );
  });
});
