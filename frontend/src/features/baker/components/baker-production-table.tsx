"use client";

import { useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import { DateInput } from "@/components/ui/date-input";
import { DashboardStepGroup } from "@/components/ui/dashboard-step-group";
import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { DashboardSearchField } from "@/components/ui/dashboard-search-field";
import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import {
  formatOrderStatusLabel,
  StatusPill,
  ticketStatusToPillVariant,
} from "@/components/ui/status-pill";
import {
  dashboardTableEmptyFiltersMessage,
  dashboardTableEmptyMessage,
  dashboardTableErrorMessage,
  DASHBOARD_TABLE_PAGE_SIZE,
} from "@/constants/dashboard-table";
import { ROUTE } from "@/constants/routes";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { formatTicketItems } from "@/lib/schemas/ticket";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { useDebouncedTableSearch } from "@/lib/hooks/use-debounced-table-search";
import { useNow } from "@/lib/hooks/use-now";
import { keepPageInRange } from "@/lib/pagination/dashboard-pagination";
import { bakeryDayOf, formatPickupDay, formatPickupWallTime } from "@/lib/validation/pickup";

import { useKitchenTickets } from "../hooks";
import { ticketItemCount, type ActiveTicketStatus } from "../schemas";

const PAGE_SIZE = DASHBOARD_TABLE_PAGE_SIZE;
const MINUTE_MS = 60_000;

const STATUS_FILTERS: Array<{ value: ActiveTicketStatus | undefined; label: string }> = [
  { value: undefined, label: "All active" },
  { value: "queued", label: "Queued" },
  { value: "in_progress", label: "In progress" },
  { value: "ready", label: "Ready" },
];

/**
 * The kitchen's queue: one row per kitchen ticket, led by its pickup time —
 * every day's, since a cake for tomorrow is made today, or one pickup day's.
 */
export function BakerProductionTable() {
  const today = bakeryDayOf(useNow(MINUTE_MS));
  const { clear, input, page, search, setInput, setPage } = useDebouncedTableSearch();
  const [status, setStatus] = useState<ActiveTicketStatus | undefined>(undefined);
  const [day, setDay] = useState<string | undefined>(undefined);
  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, status, date: day, search }),
    [page, status, day, search],
  );
  const ticketsQuery = useKitchenTickets(filter);
  const { initialLoading, refetching } = getQuerySurface(ticketsQuery);
  const tickets = ticketsQuery.data?.tickets ?? [];
  const pagination = ticketsQuery.data?.pagination;
  keepPageInRange(page, setPage, pagination?.total_pages);
  const hasActiveFilter = status !== undefined || day !== undefined || search !== "";

  function showDay(next: string | undefined): void {
    setDay(next);
    setPage(1);
  }

  return (
    <div className="dashboard-page-body">
      <div className="db-table-filters">
        <DashboardSearchField
          label="Search kitchen tickets"
          onChange={setInput}
          onClear={clear}
          placeholder="Search order code, name, item"
          value={input}
        />
        <div className="db-table-day">
          <DateInput
            id="baker-production-day"
            label="Pickup day"
            placeholder="Any day"
            value={day ?? ""}
            onChange={(next) => showDay(next || undefined)}
          />
          <DashboardStepGroup
            aria-label="Pickup days to show"
            steps={[
              { label: "Today", blockedReason: day === today ? "Already showing today" : undefined, onClick: () => showDay(today) },
              { label: "All days", blockedReason: day === undefined ? "Already showing every day" : undefined, onClick: () => showDay(undefined) },
            ]}
          />
        </div>
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
          <div aria-busy="true" className="baker-production-empty grid gap-4">
            <span className="skeleton block h-6 w-1/2" />
            <span className="skeleton block h-6 w-2/3" />
            <span className="skeleton block h-6 w-1/3" />
          </div>
        ) : ticketsQuery.isError ? (
          <div className="baker-production-empty db-table-error">
            <p className="text-error">{dashboardTableErrorMessage("kitchen tickets")}</p>
            <Button size="sm" type="button" variant="outline" onClick={() => void ticketsQuery.refetch()}>
              Try again
            </Button>
          </div>
        ) : tickets.length === 0 ? (
          <p className="baker-production-empty">
            {hasActiveFilter
              ? dashboardTableEmptyFiltersMessage("kitchen tickets")
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
                    {ticket.pickup_at ? (
                      <>
                        {day === undefined ? (
                          <span className="baker-production-row__day">{formatPickupDay(ticket.pickup_at)}</span>
                        ) : null}
                        {formatPickupWallTime(ticket.pickup_at)}
                      </>
                    ) : (
                      "—"
                    )}
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
