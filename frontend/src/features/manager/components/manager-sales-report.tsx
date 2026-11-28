"use client";

import { useState } from "react";

import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DateInput } from "@/components/ui/date-input";
import { Label } from "@/components/ui/label";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { useNow } from "@/lib/hooks/use-now";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
import { bakeryDayOf } from "@/lib/validation/pickup";

import { defaultSalesRange, rangeWithFrom, rangeWithTo } from "../lib/sales-report";
import type { SalesReportFilterInput } from "../schemas";
import { ManagerSalesCategories } from "./manager-sales-categories";
import { ManagerSalesDiscounts } from "./manager-sales-discounts";
import { ManagerSalesItems } from "./manager-sales-items";
import { ManagerSalesKpis } from "./manager-sales-kpis";
import { ManagerSalesPeriods } from "./manager-sales-periods";

const MINUTE_MS = 60_000;

const GROUPS: ReadonlyArray<{ value: SalesReportFilterInput["group"]; label: string }> = [
  { value: "day", label: "Day" },
  { value: "week", label: "Week" },
  { value: "month", label: "Month" },
];

/** What the bakery sold over a range of days: the money, the best sellers, the categories and the discounts. */
export function ManagerSalesReport() {
  const today = bakeryDayOf(useNow(MINUTE_MS));
  const [range, setRange] = useState(() => defaultSalesRange(today));

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description="Money taken and refunded in the days you pick, and what the orders paid in them sold. Up to a year at a time."
        eyebrow={DASHBOARD_PAGE_EYEBROW.reports}
        title={PAGE_TITLES.sales}
      />
      {/* The range, as the other views show their day or week: the fields left,
          how to add up right. */}
      <div className="db-table-filters">
        <div className="db-table-day db-table-day--range">
          <span className="db-range-field">
            <Label htmlFor="sales-from">From</Label>
            <DateInput
              id="sales-from"
              max={today}
              placeholder="First day"
              required
              value={range.from}
              onChange={(next) => setRange(rangeWithFrom(range, next))}
            />
          </span>
          <span className="db-range-field">
            <Label htmlFor="sales-to">To</Label>
            <DateInput
              id="sales-to"
              max={today}
              placeholder="Last day"
              required
              value={range.to}
              onChange={(next) => setRange(rangeWithTo(range, next))}
            />
          </span>
        </div>
        <DashboardFilterGroup
          aria-label="Add up by"
          onChange={(group) => setRange({ ...range, group })}
          options={GROUPS}
          value={range.group}
        />
      </div>
      <ManagerSalesKpis range={range} />
      {/* Block flow, so the sections keep their own rhythm and hairlines. */}
      <div>
        <ManagerSalesPeriods key={`${range.from}-${range.to}-${range.group}`} range={range} />
        <ManagerSalesItems range={range} />
        <ManagerSalesCategories range={range} />
        <ManagerSalesDiscounts range={range} />
      </div>
    </div>
  );
}
