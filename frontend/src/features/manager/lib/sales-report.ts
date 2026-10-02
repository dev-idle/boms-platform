import { shiftDay } from "@/lib/validation/pickup";

import type { SalesReport, SalesReportFilterInput } from "../schemas";

/** The longest range a report covers — mirrors backend `salesReportMaxDays`. */
const SALES_REPORT_MAX_DAYS = 366;

/** The last 30 bakery days up to `today`, by day. */
export function defaultSalesRange(today: string): SalesReportFilterInput {
  return { from: shiftDay(today, -29), to: today, group: "day" };
}

/** The range starting on `from`: its end moves so it stays after `from` and within a report's span. */
export function rangeWithFrom(range: SalesReportFilterInput, from: string): SalesReportFilterInput {
  const last = shiftDay(from, SALES_REPORT_MAX_DAYS - 1);
  const to = range.to < from ? from : range.to > last ? last : range.to;
  return { ...range, from, to };
}

/** The range ending on `to`: its start moves so it stays before `to` and within a report's span. */
export function rangeWithTo(range: SalesReportFilterInput, to: string): SalesReportFilterInput {
  const first = shiftDay(to, -(SALES_REPORT_MAX_DAYS - 1));
  const from = range.from > to ? to : range.from < first ? first : range.from;
  return { ...range, from, to };
}

/**
 * Why a report's download cannot start yet, or undefined when it can: the
 * report still loading or not loaded, or nothing in it to save.
 */
export function downloadBlockedReason(report: SalesReport | undefined, isError: boolean, empty: boolean): string | undefined {
  if (isError) {
    return "The report could not be loaded";
  }
  if (!report) {
    return "The report is still loading";
  }
  return empty ? "Nothing sold in these days" : undefined;
}

const DAY_FORMAT = new Intl.DateTimeFormat("en-US", { month: "short", day: "numeric", year: "numeric", timeZone: "UTC" });
const MONTH_FORMAT = new Intl.DateTimeFormat("en-US", { month: "long", year: "numeric", timeZone: "UTC" });

/** A period by its first day: "Sep 1, 2026", "Week of Sep 7, 2026" or "September 2026". */
export function formatPeriod(start: string, group: SalesReportFilterInput["group"]): string {
  const day = new Date(`${start}T00:00:00Z`);
  if (group === "month") {
    return MONTH_FORMAT.format(day);
  }
  const label = DAY_FORMAT.format(day);
  return group === "week" ? `Week of ${label}` : label;
}

/** Cents as a plain amount for a spreadsheet: 4550 → "45.50". */
export function centsToAmount(cents: number): string {
  return (cents / 100).toFixed(2);
}

/** A text cell a spreadsheet would run as a formula starts with one of these. */
const FORMULA_START = /^[=+\-@\t\r]/;
/** A cell that is a plain number, such as an amount from centsToAmount. */
const PLAIN_NUMBER = /^-?\d+(\.\d+)?$/;

/**
 * Rows as CSV (RFC 4180): numbers go as they are, a negative amount included;
 * every other cell is quoted, and one a spreadsheet would read as a formula is
 * kept as text by a leading apostrophe.
 */
export function toCsv(rows: ReadonlyArray<ReadonlyArray<string | number>>): string {
  return rows
    .map((row) =>
      row
        .map((cell) => {
          if (typeof cell === "number" || PLAIN_NUMBER.test(cell)) {
            return String(cell);
          }
          const text = FORMULA_START.test(cell) ? `'${cell}` : cell;
          return `"${text.replaceAll('"', '""')}"`;
        })
        .join(","),
    )
    .join("\r\n");
}
