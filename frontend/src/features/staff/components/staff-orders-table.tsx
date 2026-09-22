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
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { formatPickupDateTime } from "@/lib/validation/pickup";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useStaffOrders } from "../hooks";
import type { OrderStatus } from "@/lib/schemas/order";

const PAGE_SIZE = DASHBOARD_STAFF_ORDERS_PAGE_SIZE;

const STATUS_FILTERS: Array<{ value: OrderStatus | undefined; label: string }> = [
  { value: undefined, label: "All" },
  { value: "pending", label: "Pending" },
  { value: "confirmed", label: "Confirmed" },
  { value: "in_production", label: "In production" },
  { value: "ready", label: "Ready" },
  { value: "fulfilled", label: "Fulfilled" },
  { value: "cancelled", label: "Cancelled" },
];

export function StaffOrdersTable() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<OrderStatus | undefined>(undefined);
  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, status }),
    [page, status],
  );
  const ordersQuery = useStaffOrders(filter);
  const { initialLoading, refetching } = getQuerySurface(ordersQuery);
  const orders = ordersQuery.data?.orders ?? [];
  const pagination = ordersQuery.data?.pagination;
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    orders.length,
    pagination,
    PAGE_SIZE,
  );

  return (
    <div className="dashboard-page-body">
      <div className="db-table-filters db-table-filters--end">
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
        <table className="db-table db-table--staff-orders db-table--comfortable">
          <thead>
            <tr>
              <th>Order</th>
              <th>Customer</th>
              <th>Pickup</th>
              <th className="db-table-cell-numeric">Total</th>
              <th className="db-table-status">Status</th>
              <th className="db-table-detail">Detail</th>
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
              initialLoading={initialLoading}
            />
            {!initialLoading && !ordersQuery.isError && orders.length > 0
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
                    <td className="text-muted">
                      {order.pickup_at
                        ? formatPickupDateTime(order.pickup_at)
                        : "Not scheduled"}
                    </td>
                    <td className="text-tabular">
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
                          label={`Open order ${order.id}`}
                          showArrow
                          text="Open"
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
