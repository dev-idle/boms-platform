"use client";

import { useState } from "react";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { Input } from "@/components/ui/input";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { useNow } from "@/lib/hooks/use-now";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
import { bakeryDayOf, shiftDay } from "@/lib/validation/pickup";

import { weekStartOf } from "../lib/incident-week";
import { ManagerIncidentKpis } from "./manager-incident-kpis";
import { ManagerIncidentsTable } from "./manager-incidents-table";

const MINUTE_MS = 60_000;

/** What went wrong with orders, a bakery week at a time: the week's numbers, then each incident. */
export function ManagerIncidents() {
  const today = bakeryDayOf(useNow(MINUTE_MS));
  const thisWeek = weekStartOf(today);
  const [week, setWeek] = useState(thisWeek);

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <div className="db-table-day">
            <Input
              aria-label="Week starting"
              max={today}
              type="date"
              value={week}
              variant="inline"
              onChange={(event) => {
                const parsed = z.iso.date().safeParse(event.target.value);
                if (parsed.success && parsed.data <= today) {
                  setWeek(weekStartOf(parsed.data));
                }
              }}
            />
            <Button size="sm" type="button" variant="outline" onClick={() => setWeek(shiftDay(week, -7))}>
              Previous week
            </Button>
            <Button
              disabled={week === thisWeek}
              size="sm"
              title={week === thisWeek ? "Already showing this week" : undefined}
              type="button"
              variant="outline"
              onClick={() => setWeek(thisWeek)}
            >
              This week
            </Button>
            <Button
              disabled={week === thisWeek}
              size="sm"
              title={week === thisWeek ? "No later week yet" : undefined}
              type="button"
              variant="outline"
              onClick={() => setWeek(shiftDay(week, 7))}
            >
              Next week
            </Button>
          </div>
        }
        description="What the system recorded as it happened and what the counter reported, Monday to Sunday."
        eyebrow={DASHBOARD_PAGE_EYEBROW.operations}
        title={PAGE_TITLES.incidents}
      />
      <ManagerIncidentKpis week={week} />
      <div className="dashboard-page-body">
        {/* A new week starts on the first page with every type. */}
        <ManagerIncidentsTable key={week} week={week} />
      </div>
    </div>
  );
}
