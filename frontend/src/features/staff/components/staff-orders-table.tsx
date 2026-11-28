"use client";

import { useMemo, useState } from "react";

import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { DashboardSearchField } from "@/components/ui/dashboard-search-field";
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
import { DASHBOARD_TABLE_PAGE_SIZE, dashboardTableEmptyFiltersMessage } from "@/constants/dashboard-table";
import { ROUTE } from "@/constants/routes";
import { useDebouncedTableSearch } from "@/lib/hooks/use-debounced-table-search";
import { paginatedPlaceholderCountFromMeta } from "@/lib/pagination/dashboard-pagination";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { keepPageInRange } from "@/lib/pagination/dashboard-pagination";
import { formatPickupDateTime } from "@/lib/validation/pickup";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useStaffOrders } from "../hooks";
import { StaffCustomerCell } from "./staff-customer-cell";
import type { StaffOrdersStatusFilter } from "../schemas";

const PAGE_SIZE = DASHBOARD_TABLE_PAGE_SIZE;

const STATUS_FILTERS: Array<{ value: StaffOrdersStatusFilter | undefined; label: string }> = [
  { value: undefined, label: "All" },
  { value: "left_open", label: "Left open" },
  { value: "pending", label: "Pending" },
  { value: "confirmed", label: "Confirmed" },
  { value: "in_production", label: "In production" },
  { value: "ready", label: "Ready" },
  { value: "fulfilled", label: "Fulfilled" },
  { value: "cancelled", label: "Cancelled" },
  { value: "no_show", label: "No show" },
];

export function StaffOrdersTable() {
  const { clear, input, page, search, setInput, setPage } = useDebouncedTableSearch();
  const [status, setStatus] = useState<StaffOrdersStatusFilter | undefined>(undefined);
  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, status, search }),
    [page, search, status],
  );
  const ordersQuery = useStaffOrders(filter);
  const { initialLoading, refetching } = getQuerySurface(ordersQuery);
  const orders = ordersQuery.data?.orders ?? [];
  const pagination = ordersQuery.data?.pagination;
  keepPageInRange(page, setPage, pagination?.total_pages);
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    orders.length,
    pagination,
    PAGE_SIZE,
  );

  return (
    <div className="dashboard-page-body">
      <div className="db-table-filters">
        <DashboardSearchField
          label="Search orders"
          onChange={setInput}
          onClear={clear}
          placeholder="Search order code, name, email"
          value={input}
        />
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
        <table className="db-table db-table--staff-orders">
          <colgroup>
            <col className="db-table-col-code" />
            <col />
            <col className="db-table-col-pickup" />
            <col className="db-table-col-number" />
            <col className="db-table-col-status" />
            <col className="db-table-col-actions" />
          </colgroup>
          <thead>
            <tr>
              <th scope="col">Order</th>
              <th scope="col">Customer</th>
              <th scope="col">Pickup</th>
              <th scope="col" className="db-table-num">Total</th>
              <th scope="col" className="db-table-status">Status</th>
              <th scope="col" className="db-table-detail">Detail</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={6}
              emptyFilteredMessage={dashboardTableEmptyFiltersMessage("orders")}
              entityLabel="orders"
              hasActiveFilter={status !== undefined || search !== ""}
              isEmpty={orders.length === 0}
              isError={ordersQuery.isError}
              onRetry={() => void ordersQuery.refetch()}
              initialLoading={initialLoading}
            />
            {!initialLoading && !ordersQuery.isError && orders.length > 0
              ? orders.map((order) => (
                  <tr key={order.id}>
                    <td className="whitespace-nowrap">
                      <span className="text-order-code">{order.code}</span>
                    </td>
                    <td>
                      <StaffCustomerCell channel={order.channel} customer={order.customer} />
                    </td>
                    <td className="whitespace-nowrap text-muted">
                      {order.pickup_at
                        ? formatPickupDateTime(order.pickup_at)
                        : "Not scheduled"}
                    </td>
                    <td className="db-table-num">
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
                          label={`Open order ${order.code}`}
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
