"use client";

import { useState } from "react";

import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DateInput } from "@/components/ui/date-input";
import { DashboardStepGroup } from "@/components/ui/dashboard-step-group";
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
        description="What the system recorded as it happened and what the counter reported, Monday to Sunday."
        eyebrow={DASHBOARD_PAGE_EYEBROW.operations}
        title={PAGE_TITLES.incidents}
      />
      {/* The week, as the pickup schedule shows its day: the field left, the steps right. */}
      <div className="db-table-filters">
        <div className="db-table-day">
          <DateInput
            id="manager-incidents-week"
            label="Week of"
            max={today}
            placeholder="Choose a day"
            required
            value={week}
            onChange={(next) => setWeek(weekStartOf(next))}
          />
        </div>
        <DashboardStepGroup
          aria-label="Step the week"
          steps={[
            { label: "Previous week", direction: "previous", onClick: () => setWeek(shiftDay(week, -7)) },
            { label: "This week", blockedReason: week === thisWeek ? "Already showing this week" : undefined, onClick: () => setWeek(thisWeek) },
            {
              label: "Next week",
              direction: "next",
              blockedReason: week === thisWeek ? "No later week yet" : undefined,
              onClick: () => setWeek(shiftDay(week, 7)),
            },
          ]}
        />
      </div>
      <ManagerIncidentKpis week={week} />
      <div className="dashboard-page-body">
        {/* A new week starts on the first page with every type. */}
        <ManagerIncidentsTable key={week} week={week} />
      </div>
    </div>
  );
}
