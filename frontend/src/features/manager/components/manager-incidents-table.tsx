"use client";

import { useState } from "react";

import { DashboardTableDateTimeCell } from "@/components/ui/dashboard-table-datetime-cell";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { DASHBOARD_TABLE_PAGE_SIZE } from "@/constants/dashboard-table";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { INCIDENT_TYPE_LABEL, incidentTypeSchema, type IncidentType } from "@/lib/schemas/incident";
import { formatPickupDateTime } from "@/lib/validation/pickup";

import { useIncidents, useIncidentSummary } from "../hooks";
import type { ManagerIncident } from "../schemas";

const PAGE_SIZE = DASHBOARD_TABLE_PAGE_SIZE;
const COLUMN_COUNT = 4;
const ALL_TYPES = "all";

/** Who recorded an incident: the staff member, or the system as it happened. */
function recordedBy(incident: ManagerIncident): string {
  if (!incident.actor_name) {
    return "Recorded automatically";
  }
  return incident.source === "manual" ? `Reported by ${incident.actor_name}` : `By ${incident.actor_name}`;
}

type ManagerIncidentsTableProps = {
  /** The Monday (YYYY-MM-DD) the week starts on. */
  week: string;
};

/** A week's incidents, latest first, narrowed to one type; each type shows how many the week holds. */
export function ManagerIncidentsTable({ week }: ManagerIncidentsTableProps) {
  const [page, setPage] = useState(1);
  const [type, setType] = useState<IncidentType | undefined>(undefined);
  const incidentsQuery = useIncidents({ week, type, page, page_size: PAGE_SIZE });
  const summary = useIncidentSummary(week).data;
  const { initialLoading, refetching } = getQuerySurface(incidentsQuery);
  const incidents = incidentsQuery.data?.incidents ?? [];
  const pagination = incidentsQuery.data?.pagination;
  const counted = (label: string, count: number | undefined) => (summary ? `${label} (${count ?? 0})` : label);

  return (
    <>
      <div className="db-table-filters db-table-filters--end">
        <Label className="sr-only" htmlFor="incident-type">
          Incident type
        </Label>
        <Select
          className="field-chrome--inline"
          id="incident-type"
          value={type ?? ALL_TYPES}
          onChange={(event) => {
            const parsed = incidentTypeSchema.safeParse(event.target.value);
            setType(parsed.success ? parsed.data : undefined);
            setPage(1);
          }}
        >
          <option value={ALL_TYPES}>
            {counted("All incidents", summary?.types.reduce((sum, { count }) => sum + count, 0))}
          </option>
          {incidentTypeSchema.options.map((option) => (
            <option key={option} value={option}>
              {counted(INCIDENT_TYPE_LABEL[option], summary?.types.find((entry) => entry.type === option)?.count)}
            </option>
          ))}
        </Select>
      </div>

      <DashboardTableWrap refetching={refetching}>
        <table className="db-table db-table--incidents">
          <colgroup>
            <col className="db-table-col-incident" />
            <col className="db-table-col-order" />
            <col />
            <col className="db-table-col-datetime" />
          </colgroup>
          <thead>
            <tr>
              <th>Incident</th>
              <th>Order</th>
              <th>Note</th>
              <th>Recorded</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={COLUMN_COUNT}
              emptyFilteredMessage="No incident of this type in the week."
              emptyMessage="No incident in the week."
              entityLabel="incidents"
              hasActiveFilter={type !== undefined}
              isEmpty={incidents.length === 0}
              isError={incidentsQuery.isError}
              initialLoading={initialLoading}
            />
            {!initialLoading && !incidentsQuery.isError
              ? incidents.map((incident) => (
                  <tr key={incident.id}>
                    <td>
                      <div className="db-table-stacked-cell min-w-0">
                        <span className="db-table-cell-primary">{INCIDENT_TYPE_LABEL[incident.type]}</span>
                        <span className="truncate text-caption-dashboard text-muted">{recordedBy(incident)}</span>
                      </div>
                    </td>
                    <td>
                      <div className="db-table-stacked-cell min-w-0">
                        <span className="text-order-code">{incident.order_code}</span>
                        {incident.pickup_at ? (
                          <span className="text-caption-dashboard text-muted">
                            Pickup {formatPickupDateTime(incident.pickup_at)}
                          </span>
                        ) : null}
                      </div>
                    </td>
                    <td>
                      {incident.note ? (
                        <p className="whitespace-pre-line break-words">{incident.note}</p>
                      ) : (
                        <span className="text-muted">—</span>
                      )}
                    </td>
                    <DashboardTableDateTimeCell iso={incident.created_at} />
                  </tr>
                ))
              : null}
          </tbody>
        </table>
        <DashboardTablePagination
          disabled={refetching}
          itemLabel="incidents"
          onPageChange={setPage}
          page={pagination?.page ?? page}
          pageSize={pagination?.page_size ?? PAGE_SIZE}
          totalItems={pagination?.total ?? incidents.length}
          totalPages={pagination?.total_pages ?? 1}
        />
      </DashboardTableWrap>
    </>
  );
}
