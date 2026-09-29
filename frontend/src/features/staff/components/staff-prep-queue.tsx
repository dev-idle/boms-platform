"use client";

import { useMemo, useState } from "react";

import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { DashboardTableActionButton } from "@/components/ui/dashboard-table-action-button";
import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { DashboardTablePagePlaceholders } from "@/components/ui/dashboard-table-page-placeholders";
import { DashboardTableRowActions } from "@/components/ui/dashboard-table-actions";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import {
  formatOrderStatusLabel,
  StatusPill,
  ticketStatusToPillVariant,
} from "@/components/ui/status-pill";
import { DASHBOARD_STAFF_ORDERS_PAGE_SIZE } from "@/constants/dashboard-table";
import { ROUTE } from "@/constants/routes";
import { paginatedPlaceholderCountFromMeta } from "@/lib/pagination/dashboard-pagination";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import {
  formatTicketItems,
  nextTicketAction,
  type ActiveTicketStatus,
} from "@/lib/schemas/ticket";
import { formatPickupDateTime } from "@/lib/validation/pickup";

import { useCounterTickets, usePatchCounterTicketStatus } from "../hooks";

const PAGE_SIZE = DASHBOARD_STAFF_ORDERS_PAGE_SIZE;

const STATUS_FILTERS: Array<{ value: ActiveTicketStatus | undefined; label: string }> = [
  { value: undefined, label: "All active" },
  { value: "queued", label: "Queued" },
  { value: "in_progress", label: "In progress" },
  { value: "ready", label: "Ready" },
];

/** The counter's prep queue: its tickets of accepted orders, by pickup time. */
export function StaffPrepQueue() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<ActiveTicketStatus | undefined>(undefined);
  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, status }),
    [page, status],
  );
  const ticketsQuery = useCounterTickets(filter);
  const patchStatus = usePatchCounterTicketStatus();
  const { initialLoading, refetching } = getQuerySurface(ticketsQuery);
  const tickets = ticketsQuery.data?.tickets ?? [];
  const pagination = ticketsQuery.data?.pagination;
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    tickets.length,
    pagination,
    PAGE_SIZE,
  );

  return (
    <div className="dashboard-page-body">
      <div className="db-table-filters db-table-filters--end">
        <DashboardFilterGroup
          aria-label="Filter by ticket status"
          onChange={(next) => {
            setStatus(next);
            setPage(1);
          }}
          options={STATUS_FILTERS}
          value={status}
        />
      </div>

      <DashboardTableWrap refetching={refetching}>
        <table className="db-table db-table--prep-queue">
          <colgroup>
            <col className="db-table-col-datetime" />
            <col />
            <col />
            <col className="db-table-col-status" />
            <col className="db-table-col-actions" />
          </colgroup>
          <thead>
            <tr>
              <th>Pickup</th>
              <th>Order</th>
              <th>Items</th>
              <th className="db-table-status">Status</th>
              <th className="db-table-detail">Actions</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={5}
              emptyFilteredMessage="No counter tickets match this filter."
              entityLabel="counter tickets"
              hasActiveFilter={status !== undefined}
              isEmpty={tickets.length === 0}
              isError={ticketsQuery.isError}
              initialLoading={initialLoading}
            />
            {!initialLoading && !ticketsQuery.isError && tickets.length > 0
              ? tickets.map((ticket) => {
                  const action = nextTicketAction(ticket.status);
                  // A ready ticket keeps the slot, greyed: the counter hands it over from the order.
                  const actionText = action?.label ?? "Mark ready";
                  const isUpdating =
                    patchStatus.isPending && patchStatus.variables?.ticketId === ticket.id;
                  return (
                    <tr key={ticket.id}>
                      <td className="text-muted">
                        {ticket.pickup_at
                          ? formatPickupDateTime(ticket.pickup_at)
                          : "Not scheduled"}
                      </td>
                      <td>
                        <div className="db-table-stacked-cell min-w-0">
                          <span className="text-order-code">{ticket.order_code}</span>
                          <span className="truncate text-caption-dashboard text-muted">
                            {ticket.customer.display_name || "Customer"}
                          </span>
                        </div>
                      </td>
                      <td>{formatTicketItems(ticket.items)}</td>
                      <td className="db-table-status">
                        <StatusPill
                          label={formatOrderStatusLabel(ticket.status)}
                          variant={ticketStatusToPillVariant(ticket.status)}
                        />
                      </td>
                      <td className="db-table-detail">
                        <DashboardTableRowActions>
                          <DashboardTableActionButton
                            blockedReason={
                              isUpdating
                                ? "Updating the ticket"
                                : action
                                  ? undefined
                                  : "The ticket is ready for pickup"
                            }
                            label={`${actionText} for order ${ticket.order_code}`}
                            onClick={() => {
                              if (action) {
                                patchStatus.mutate({ ticketId: ticket.id, status: action.status });
                              }
                            }}
                            text={actionText}
                            tone="accent"
                          />
                          <DashboardTableActionLink
                            href={ROUTE.staff.orderDetail(ticket.order_id)}
                            label={`Open order ${ticket.order_code}`}
                            showArrow
                            text="Order"
                          />
                        </DashboardTableRowActions>
                      </td>
                    </tr>
                  );
                })
              : null}
            <DashboardTablePagePlaceholders
              columnCount={5}
              count={pagePlaceholderCount}
            />
          </tbody>
        </table>
        <DashboardTablePagination
          disabled={refetching}
          itemLabel="tickets"
          onPageChange={setPage}
          page={pagination?.page ?? page}
          pageSize={pagination?.page_size ?? PAGE_SIZE}
          totalItems={pagination?.total ?? tickets.length}
          totalPages={pagination?.total_pages ?? 1}
        />
      </DashboardTableWrap>
    </div>
  );
}
