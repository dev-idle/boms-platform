"use client";

import { useMemo, useState } from "react";

import { BusyIndicator } from "@/components/ui/loading-state";
import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import {
  formatOrderStatusLabel,
  StatusPill,
  ticketStatusToPillVariant,
} from "@/components/ui/status-pill";
import {
  dashboardTableEmptyMessage,
  dashboardTableErrorMessage,
  DASHBOARD_STAFF_ORDERS_PAGE_SIZE,
} from "@/constants/dashboard-table";
import { ROUTE } from "@/constants/routes";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import {
  formatTicketItems,
  ticketItemCount,
  type ActiveTicketStatus,
} from "@/lib/schemas/ticket";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { formatPickupWallTime } from "@/lib/validation/pickup";

import { useKitchenTickets } from "../hooks";

const PAGE_SIZE = DASHBOARD_STAFF_ORDERS_PAGE_SIZE;

const STATUS_FILTERS: Array<{ value: ActiveTicketStatus | undefined; label: string }> = [
  { value: undefined, label: "All active" },
  { value: "queued", label: "Queued" },
  { value: "in_progress", label: "In progress" },
  { value: "ready", label: "Ready" },
];

/** The kitchen's queue: one row per kitchen ticket, led by its pickup time. */
export function BakerProductionTable() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<ActiveTicketStatus | undefined>(undefined);
  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, status }),
    [page, status],
  );
  const ticketsQuery = useKitchenTickets(filter);
  const { initialLoading, refetching } = getQuerySurface(ticketsQuery);
  const tickets = ticketsQuery.data?.tickets ?? [];
  const pagination = ticketsQuery.data?.pagination;
  const hasActiveFilter = status !== undefined;

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
        {initialLoading ? (
          <div className="baker-production-empty" role="status">
            <BusyIndicator />
          </div>
        ) : ticketsQuery.isError ? (
          <p className="baker-production-empty baker-production-empty--error">
            {dashboardTableErrorMessage("kitchen tickets")}
          </p>
        ) : tickets.length === 0 ? (
          <p className="baker-production-empty">
            {hasActiveFilter
              ? "No kitchen tickets match this filter."
              : dashboardTableEmptyMessage("kitchen tickets")}
          </p>
        ) : (
          <>
            {/* A header strip, not <thead>: the list stays a list. */}
            <div aria-hidden className="baker-production-labels">
              <span>Pickup</span>
              <span>Order</span>
              <span className="baker-production-labels__count">Items</span>
              <span>Status</span>
              <span />
            </div>
            <ul className="baker-production-list">
              {tickets.map((ticket) => (
                <li key={ticket.id} className="baker-production-row">
                  <p className="baker-production-row__time">
                    {ticket.pickup_at ? formatPickupWallTime(ticket.pickup_at) : "—"}
                  </p>
                  <div className="baker-production-row__customer">
                    <p className="baker-production-row__name">
                      {ticket.customer.display_name || "Customer"}
                    </p>
                    <p className="baker-production-row__summary">
                      {formatTicketItems(ticket.items)}
                    </p>
                    <p className="baker-production-row__code">{ticket.order_code}</p>
                  </div>
                  <span className="baker-production-row__count">
                    {ticketItemCount(ticket.items)}
                    <span className="sr-only"> items</span>
                  </span>
                  <span className="baker-production-row__status">
                    <StatusPill
                      label={formatOrderStatusLabel(ticket.status)}
                      variant={ticketStatusToPillVariant(ticket.status)}
                    />
                  </span>
                  <span className="baker-production-row__open">
                    <DashboardTableActionLink
                      href={ROUTE.baker.productionDetail(ticket.id)}
                      label={`Open the kitchen ticket for order ${ticket.order_code}`}
                      showArrow
                      text="Open"
                    />
                  </span>
                </li>
              ))}
            </ul>
          </>
        )}
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
