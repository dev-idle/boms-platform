import { describe, expect, it } from "vitest";

import {
  centsToAmount,
  defaultSalesRange,
  downloadBlockedReason,
  formatPeriod,
  rangeWithFrom,
  rangeWithTo,
  toCsv,
} from "./sales-report";
import { salesReportSchema } from "../schemas";

describe("defaultSalesRange", () => {
  it("is the last 30 days by day", () => {
    expect(defaultSalesRange("2026-10-02")).toEqual({ from: "2026-09-03", to: "2026-10-02", group: "day" });
  });
});

describe("rangeWithFrom and rangeWithTo", () => {
  const range = { from: "2026-09-01", to: "2026-09-30", group: "week" as const };

  it("keep the other end when the range still holds", () => {
    expect(rangeWithFrom(range, "2026-09-10")).toEqual({ ...range, from: "2026-09-10" });
    expect(rangeWithTo(range, "2026-09-20")).toEqual({ ...range, to: "2026-09-20" });
  });

  it("move the other end past the one picked", () => {
    expect(rangeWithFrom(range, "2026-10-01")).toEqual({ ...range, from: "2026-10-01", to: "2026-10-01" });
    expect(rangeWithTo(range, "2026-08-15")).toEqual({ ...range, from: "2026-08-15", to: "2026-08-15" });
  });

  it("hold a range to 366 days", () => {
    expect(rangeWithFrom({ ...range, to: "2026-09-30" }, "2025-01-01")).toEqual({
      ...range,
      from: "2025-01-01",
      to: "2026-01-01",
    });
    expect(rangeWithTo({ ...range, from: "2025-01-01" }, "2026-09-30")).toEqual({
      ...range,
      from: "2025-09-30",
      to: "2026-09-30",
    });
  });
});

describe("downloadBlockedReason", () => {
  const report = salesReportSchema.parse({
    from: "2026-09-01",
    to: "2026-09-01",
    group: "day",
    totals: { orders: 0, gross_cents: 0, refunds_cents: 0, net_cents: 0, average_order_cents: null },
    periods: [],
    items: [],
    categories: [],
    discounts: { orders: 0, discounted_orders: 0, discount_cents: 0, codes: [] },
    production: { orders: 0, average_minutes: null },
  });

  it("gives the true reason a download cannot start", () => {
    expect(downloadBlockedReason(undefined, true, true)).toBe("The report could not be loaded");
    expect(downloadBlockedReason(undefined, false, true)).toBe("The report is still loading");
    expect(downloadBlockedReason(report, false, true)).toBe("Nothing sold in these days");
    expect(downloadBlockedReason(report, false, false)).toBeUndefined();
  });
});

describe("formatPeriod", () => {
  it("names a day, a week and a month", () => {
    expect(formatPeriod("2026-09-01", "day")).toBe("Sep 1, 2026");
    expect(formatPeriod("2026-09-07", "week")).toBe("Week of Sep 7, 2026");
    expect(formatPeriod("2026-09-01", "month")).toBe("September 2026");
  });
});

describe("centsToAmount", () => {
  it("writes cents as a plain amount", () => {
    expect(centsToAmount(4550)).toBe("45.50");
    expect(centsToAmount(-500)).toBe("-5.00");
    expect(centsToAmount(0)).toBe("0.00");
  });
});

describe("toCsv", () => {
  it("quotes text, leaves numbers and ends rows with CRLF", () => {
    expect(
      toCsv([
        ["Item", "Quantity"],
        ['Cake "Mai"', 2],
      ]),
    ).toBe('"Item","Quantity"\r\n"Cake ""Mai""",2');
  });

  it("writes an amount as a number, a negative one included", () => {
    expect(toCsv([["2026-09-07", 1, centsToAmount(-4600), "12.50"]])).toBe('"2026-09-07",1,-46.00,12.50');
  });

  it("keeps a cell that reads as a formula as text", () => {
    expect(toCsv([["=HYPERLINK(1)", "+1", "-x", "@sum", "Plain"]])).toBe(
      `"'=HYPERLINK(1)","'+1","'-x","'@sum","Plain"`,
    );
  });
});
