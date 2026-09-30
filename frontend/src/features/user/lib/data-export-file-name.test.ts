import { describe, expect, it } from "vitest";

import { dataExportFileName } from "./data-export-file-name";

describe("dataExportFileName", () => {
  it("names the bakery and the day the export was taken", () => {
    expect(dataExportFileName(new Date("2026-09-29T08:15:00Z"))).toBe("choux-my-data-2026-09-29.json");
  });

  it("takes the day in the bakery's time zone, not UTC", () => {
    // 06:00 on 30 September in Vietnam is still 29 September in UTC.
    expect(dataExportFileName(new Date("2026-09-29T23:00:00Z"))).toBe("choux-my-data-2026-09-30.json");
  });
});
