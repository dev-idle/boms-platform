"use client";

import { useState } from "react";
import { z } from "zod";

import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { Input } from "@/components/ui/input";
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
        actions={
          <div className="db-table-day db-table-day--range">
            <span className="db-range-field">
              <Label htmlFor="sales-from">From</Label>
              <Input
                id="sales-from"
                max={today}
                type="date"
                value={range.from}
                variant="inline"
                onChange={(event) => {
                  const parsed = z.iso.date().safeParse(event.target.value);
                  if (parsed.success && parsed.data <= today) {
                    setRange(rangeWithFrom(range, parsed.data));
                  }
                }}
              />
            </span>
            <span className="db-range-field">
              <Label htmlFor="sales-to">To</Label>
              <Input
                id="sales-to"
                max={today}
                type="date"
                value={range.to}
                variant="inline"
                onChange={(event) => {
                  const parsed = z.iso.date().safeParse(event.target.value);
                  if (parsed.success && parsed.data <= today) {
                    setRange(rangeWithTo(range, parsed.data));
                  }
                }}
              />
            </span>
            <DashboardFilterGroup
              aria-label="Add up by"
              onChange={(group) => setRange({ ...range, group })}
              options={GROUPS}
              value={range.group}
            />
          </div>
        }
        description="Money taken and refunded in the days you pick, and what the orders paid in them sold. Up to a year at a time."
        eyebrow={DASHBOARD_PAGE_EYEBROW.reports}
        title={PAGE_TITLES.sales}
      />
      <ManagerSalesKpis range={range} />
      {/* Block flow, so the sections keep their own rhythm and hairlines. */}
      <div>
        <ManagerSalesPeriods range={range} />
        <ManagerSalesItems range={range} />
        <ManagerSalesCategories range={range} />
        <ManagerSalesDiscounts range={range} />
      </div>
    </div>
  );
}
