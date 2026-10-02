"use client";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { formatOrderStatusLabel, orderStatusToPillVariant, StatusPill } from "@/components/ui/status-pill";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { orderStatusSchema } from "@/lib/schemas/order";
import { STATION_LABEL, stationSchema } from "@/lib/schemas/ticket";

import { useOperations } from "../hooks";

/** The statuses the bakery still works on, in the order an order moves through them. */
const ACTIVE_STATUSES = orderStatusSchema.extract(["pending", "confirmed", "in_production", "ready"]).options;

/** The open orders by status, then each station's tickets: where the work sits now. */
export function ManagerOperationsWork() {
  const operationsQuery = useOperations();
  const { initialLoading, refetching } = getQuerySurface(operationsQuery);
  const operations = operationsQuery.data;
  const loaded = operations !== undefined && !initialLoading && !operationsQuery.isError;

  return (
    <>
      <DashboardProfileSection
        description="Every order still to make or hand over, whatever its pickup day."
        id="manager-operations-orders"
        title="Orders"
        variant="plain"
      >
        <DashboardTableWrap refetching={refetching}>
          <table className="db-table">
            <colgroup>
              <col />
              <col className="db-table-col-number" />
            </colgroup>
            <thead>
              <tr>
                <th>Status</th>
                <th className="db-table-num">Orders</th>
              </tr>
            </thead>
            <tbody>
              <DashboardTableStateRows
                columnCount={2}
                entityLabel="orders"
                isEmpty={false}
                isError={operationsQuery.isError}
                initialLoading={initialLoading}
              />
              {loaded
                ? ACTIVE_STATUSES.map((status) => (
                    <tr key={status}>
                      <td>
                        <StatusPill label={formatOrderStatusLabel(status)} variant={orderStatusToPillVariant(status)} />
                      </td>
                      <td className="db-table-num">{operations.orders[status]}</td>
                    </tr>
                  ))
                : null}
            </tbody>
          </table>
        </DashboardTableWrap>
      </DashboardProfileSection>
      <DashboardProfileSection
        description="The tickets of those orders at each station: waiting, being made, and done."
        id="manager-operations-stations"
        title="Stations"
        variant="plain"
      >
        <DashboardTableWrap refetching={refetching}>
          <table className="db-table db-table--figures-3">
            <colgroup>
              <col />
              <col className="db-table-col-number" />
              <col className="db-table-col-number" />
              <col className="db-table-col-number" />
            </colgroup>
            <thead>
              <tr>
                <th>Station</th>
                <th className="db-table-num">Queued</th>
                <th className="db-table-num">In progress</th>
                <th className="db-table-num">Ready</th>
              </tr>
            </thead>
            <tbody>
              <DashboardTableStateRows
                columnCount={4}
                entityLabel="stations"
                isEmpty={false}
                isError={operationsQuery.isError}
                initialLoading={initialLoading}
              />
              {loaded
                ? stationSchema.options.map((station) => {
                    const load = operations.stations.find((entry) => entry.station === station);
                    return (
                      <tr key={station}>
                        <td className="db-table-cell-primary">{STATION_LABEL[station]}</td>
                        <td className="db-table-num">{load?.queued ?? 0}</td>
                        <td className="db-table-num">{load?.in_progress ?? 0}</td>
                        <td className="db-table-num">{load?.ready ?? 0}</td>
                      </tr>
                    );
                  })
                : null}
            </tbody>
          </table>
        </DashboardTableWrap>
      </DashboardProfileSection>
    </>
  );
}
