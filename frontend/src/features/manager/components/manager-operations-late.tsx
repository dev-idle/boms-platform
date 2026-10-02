"use client";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { formatOrderStatusLabel, orderStatusToPillVariant, StatusPill } from "@/components/ui/status-pill";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { formatPickupDateTime } from "@/lib/validation/pickup";

import { useOperations } from "../hooks";
import type { Operations } from "../schemas";

const COLUMN_COUNT = 3;

/** Today's pickups in a sentence, and how many of the late ones the table lists. */
function pickupsLead({ pickups, late }: Operations): string {
  const today = `Today: ${pickups.due} due — ${pickups.collected} collected, ${pickups.to_collect} still to collect, ${pickups.missed} not collected.`;
  return late.total > late.orders.length
    ? `${today} The table lists the ${late.orders.length} of ${late.total} waiting longest.`
    : today;
}

/** Today's pickups, and the orders still to collect whose pickup slot has ended, the longest waiting first. */
export function ManagerOperationsLate() {
  const operationsQuery = useOperations();
  const { initialLoading, refetching } = getQuerySurface(operationsQuery);
  const operations = operationsQuery.data;
  const late = operations?.late.orders ?? [];

  return (
    <DashboardProfileSection
      description={operations ? pickupsLead(operations) : "Today's pickups, and the orders whose pickup slot has ended uncollected."}
      id="manager-operations-late"
      title="Late pickups"
      variant="plain"
    >
      <DashboardTableWrap refetching={refetching}>
        <table className="db-table db-table--late">
          <colgroup>
            <col className="db-table-col-code" />
            <col />
            <col className="db-table-col-status" />
          </colgroup>
          <thead>
            <tr>
              <th>Order</th>
              <th>Pickup</th>
              <th className="db-table-status">Status</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={COLUMN_COUNT}
              emptyMessage="Nothing is late."
              entityLabel="late pickups"
              isEmpty={late.length === 0}
              isError={operationsQuery.isError}
              initialLoading={initialLoading}
            />
            {!initialLoading && !operationsQuery.isError
              ? late.map((order) => (
                  <tr key={order.code}>
                    <td>
                      <span className="text-order-code">{order.code}</span>
                    </td>
                    <td className="text-muted">{formatPickupDateTime(order.pickup_at)}</td>
                    <td className="db-table-status">
                      <StatusPill label={formatOrderStatusLabel(order.status)} variant={orderStatusToPillVariant(order.status)} />
                    </td>
                  </tr>
                ))
              : null}
          </tbody>
        </table>
      </DashboardTableWrap>
    </DashboardProfileSection>
  );
}
