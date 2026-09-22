"use client";

import { useMemo, useState } from "react";

import { BusyIndicator } from "@/components/ui/loading-state";
import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import {
  formatOrderStatusLabel,
  orderStatusToPillVariant,
  StatusPill,
} from "@/components/ui/status-pill";
import {
  dashboardTableEmptyMessage,
  dashboardTableErrorMessage,
  DASHBOARD_STAFF_ORDERS_PAGE_SIZE,
} from "@/constants/dashboard-table";
import { ROUTE } from "@/constants/routes";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { formatPickupWallTime } from "@/lib/validation/pickup";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useBakerProductionOrders } from "../hooks";
import type { BakerOrderStatus } from "../schemas";

const PAGE_SIZE = DASHBOARD_STAFF_ORDERS_PAGE_SIZE;

const STATUS_FILTERS: Array<{ value: BakerOrderStatus | undefined; label: string }> = [
  { value: undefined, label: "All active" },
  { value: "confirmed", label: "Confirmed" },
  { value: "in_production", label: "In production" },
  { value: "ready", label: "Ready" },
];

function bakerOrderItemSummary(itemCount: number): string {
  return itemCount === 1 ? "1 item" : `${itemCount} items`;
}

export function BakerProductionTable() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<BakerOrderStatus | undefined>(undefined);
  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, status }),
    [page, status],
  );
  const ordersQuery = useBakerProductionOrders(filter);
  const { initialLoading, refetching } = getQuerySurface(ordersQuery);
  const orders = ordersQuery.data?.orders ?? [];
  const pagination = ordersQuery.data?.pagination;
  const hasActiveFilter = status !== undefined;

  return (
    <div className="dashboard-page-body">
      <div className="db-table-filters db-table-filters--end">
        <DashboardFilterGroup
          aria-label="Filter by production status"
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
        ) : ordersQuery.isError ? (
          <p className="baker-production-empty baker-production-empty--error">
            {dashboardTableErrorMessage("production orders")}
          </p>
        ) : orders.length === 0 ? (
          <p className="baker-production-empty">
            {hasActiveFilter
              ? "No production orders match this filter."
              : dashboardTableEmptyMessage("production orders")}
          </p>
        ) : (
          <>
            {/* A header strip, not <thead>: the list stays a list. */}
            <div aria-hidden className="baker-production-labels">
              <span>Pickup</span>
              <span>Order</span>
              <span className="baker-production-labels__total">Total</span>
              <span>Status</span>
              <span />
            </div>
            <ul className="baker-production-list">
              {orders.map((order) => {
                const customerName =
                  order.customer.display_name ?? order.customer.email;
                return (
                  <li key={order.id} className="baker-production-row">
                    <p className="baker-production-row__time">
                      {order.pickup_at
                        ? formatPickupWallTime(order.pickup_at)
                        : "—"}
                    </p>
                    <div className="baker-production-row__customer">
                      <p className="baker-production-row__name">{customerName}</p>
                      <p className="baker-production-row__summary">
                        {bakerOrderItemSummary(order.item_count)}
                      </p>
                      <p className="baker-production-row__code">
                        {order.id.slice(0, 8)}
                      </p>
                    </div>
                    <span className="baker-production-row__total">
                      {formatPriceCents(order.total_cents)}
                    </span>
                    <span className="baker-production-row__status">
                      <StatusPill
                        label={formatOrderStatusLabel(order.status)}
                        variant={orderStatusToPillVariant(order.status)}
                      />
                    </span>
                    <span className="baker-production-row__open">
                      <DashboardTableActionLink
                        href={ROUTE.baker.productionDetail(order.id)}
                        label={`Open production order ${order.id}`}
                        showArrow
                        text="Open"
                      />
                    </span>
                  </li>
                );
              })}
            </ul>
          </>
        )}
        <DashboardTablePagination
          disabled={refetching}
          itemLabel="orders"
          onPageChange={setPage}
          page={pagination?.page ?? page}
          pageSize={pagination?.page_size ?? PAGE_SIZE}
          totalItems={pagination?.total ?? orders.length}
          totalPages={pagination?.total_pages ?? 1}
        />
      </DashboardTableWrap>
    </div>
  );
}
