"use client";

import { useMemo, useState } from "react";

import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { DashboardTablePagePlaceholders } from "@/components/ui/dashboard-table-page-placeholders";
import { DashboardTableRowActions } from "@/components/ui/dashboard-table-actions";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import {
  formatOrderStatusLabel,
  orderStatusToPillVariant,
  StatusPill,
} from "@/components/ui/status-pill";
import { DASHBOARD_STAFF_ORDERS_PAGE_SIZE } from "@/constants/dashboard-table";
import { ROUTE } from "@/constants/routes";
import { paginatedPlaceholderCountFromMeta } from "@/lib/pagination/dashboard-pagination";
import {
  isInitialQueryLoad,
  isQueryRefetching,
} from "@/lib/react-query/query-surface";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useStaffOrders } from "../hooks";
import type { OrderStatus } from "../schemas";

const PAGE_SIZE = DASHBOARD_STAFF_ORDERS_PAGE_SIZE;

const STATUS_FILTERS: Array<{ value: OrderStatus | undefined; label: string }> = [
  { value: undefined, label: "All" },
  { value: "pending", label: "Pending" },
  { value: "confirmed", label: "Confirmed" },
  { value: "fulfilled", label: "Fulfilled" },
  { value: "cancelled", label: "Cancelled" },
];

export function StaffOrdersTable() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<OrderStatus | undefined>("pending");
  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, status }),
    [page, status],
  );
  const ordersQuery = useStaffOrders(filter);
  const orders = ordersQuery.data?.orders ?? [];
  const pagination = ordersQuery.data?.pagination;
  const initialLoad = isInitialQueryLoad(ordersQuery.isPending, ordersQuery.data);
  const refetching = isQueryRefetching(
    ordersQuery.isFetching,
    ordersQuery.isPending,
    ordersQuery.data,
  );
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    orders.length,
    pagination,
    PAGE_SIZE,
  );

  return (
    <div className="dashboard-page-body">
      <div className="db-table-filters">
        <DashboardFilterGroup
          aria-label="Filter by status"
          onChange={(next) => {
            setStatus(next);
            setPage(1);
          }}
          options={STATUS_FILTERS}
          value={status}
        />
      </div>

      <DashboardTableWrap refetching={refetching}>
        <table className="db-table db-table--catalog db-table--relaxed">
          <thead>
            <tr>
              <th>Order</th>
              <th>Customer</th>
              <th>Placed</th>
              <th>Total</th>
              <th className="db-table-status">Status</th>
              <th className="db-table-detail">Actions</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={6}
              emptyFilteredMessage="No orders match this filter."
              entityLabel="orders"
              hasActiveFilter={status !== undefined}
              isEmpty={orders.length === 0}
              isError={ordersQuery.isError}
              isInitialLoad={initialLoad}
            />
            {!initialLoad && !ordersQuery.isError && orders.length > 0
              ? orders.map((order) => (
                  <tr key={order.id}>
                    <td>
                      <span className="text-order-code">{order.id.slice(0, 8)}</span>
                    </td>
                    <td>
                      <div className="db-table-stacked-cell min-w-0">
                        <span className="db-table-cell-primary truncate">
                          {order.customer.display_name ?? order.customer.email}
                        </span>
                        {order.customer.display_name ? (
                          <span className="truncate text-caption-dashboard text-muted">
                            {order.customer.email}
                          </span>
                        ) : null}
                      </div>
                    </td>
                    <td className="text-muted">{formatDateTime(order.created_at)}</td>
                    <td className="db-table-cell-primary text-tabular">
                      {formatPriceCents(order.total_cents)}
                    </td>
                    <td className="db-table-status">
                      <StatusPill
                        label={formatOrderStatusLabel(order.status)}
                        variant={orderStatusToPillVariant(order.status)}
                      />
                    </td>
                    <td className="db-table-detail">
                      <DashboardTableRowActions>
                        <DashboardTableActionLink
                          href={ROUTE.staff.orderDetail(order.id)}
                          label={`View order ${order.id}`}
                        />
                      </DashboardTableRowActions>
                    </td>
                  </tr>
                ))
              : null}
            <DashboardTablePagePlaceholders
              columnCount={6}
              count={pagePlaceholderCount}
            />
          </tbody>
        </table>
        <DashboardTablePagination
          disabled={ordersQuery.isFetching}
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
