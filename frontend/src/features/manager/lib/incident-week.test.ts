import { describe, expect, it } from "vitest";

import { weekStartOf } from "./incident-week";

describe("weekStartOf", () => {
  it("starts a week on its Monday", () => {
    expect(weekStartOf("2026-09-28")).toBe("2026-09-28");
    expect(weekStartOf("2026-10-02")).toBe("2026-09-28");
    expect(weekStartOf("2026-10-04"), "a Sunday closes its week").toBe("2026-09-28");
  });

  it("reaches back across months and years", () => {
    expect(weekStartOf("2026-10-01")).toBe("2026-09-28");
    expect(weekStartOf("2027-01-02")).toBe("2026-12-28");
  });
});
