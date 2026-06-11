"use client";

import Link from "next/link";
import { useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import {
  formatOrderStatusLabel,
  orderStatusToPillVariant,
  StatusPill,
} from "@/components/ui/status-pill";
import { ROUTE } from "@/constants/routes";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useStaffOrders } from "../hooks";
import type { OrderStatus } from "../schemas";

const PAGE_SIZE = 20;

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

  if (ordersQuery.isPending) {
    return <p className="text-sm text-muted">Loading orders…</p>;
  }

  if (ordersQuery.isError) {
    return <p className="text-sm text-error">Failed to load orders.</p>;
  }

  const orders = ordersQuery.data?.orders ?? [];
  const pagination = ordersQuery.data?.pagination;

  return (
    <div className="space-y-4">
      <div className="db-filter-row">
        <span className="db-filter-label">Status</span>
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

      {orders.length === 0 ? (
        <p className="text-sm text-muted">No orders match this filter.</p>
      ) : (
        <div className="db-table-wrap">
          <table className="db-table">
            <thead>
              <tr>
                <th>Order</th>
                <th>Customer</th>
                <th>Placed</th>
                <th>Total</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {orders.map((order) => (
                <tr key={order.id}>
                  <td>
                    <span className="text-order-code">{order.id.slice(0, 8)}</span>
                  </td>
                  <td>
                    <div className="min-w-0">
                      <p className="truncate font-medium text-ink">
                        {order.customer.display_name ?? order.customer.email}
                      </p>
                      {order.customer.display_name ? (
                        <p className="truncate text-caption-dashboard">
                          {order.customer.email}
                        </p>
                      ) : null}
                    </div>
                  </td>
                  <td className="text-ink-2">{formatDateTime(order.created_at)}</td>
                  <td className="text-tabular font-medium">
                    {formatPriceCents(order.total_cents)}
                  </td>
                  <td>
                    <StatusPill
                      label={formatOrderStatusLabel(order.status)}
                      variant={orderStatusToPillVariant(order.status)}
                    />
                  </td>
                  <td>
                    <Link href={ROUTE.staff.orderDetail(order.id)}>
                      <Button size="sm" type="button" variant="outline">
                        View
                      </Button>
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {pagination ? (
            <DashboardTablePagination
              disabled={ordersQuery.isFetching}
              onPageChange={setPage}
              page={pagination.page}
              pageSize={pagination.page_size}
              totalItems={pagination.total}
              totalPages={pagination.total_pages}
            />
          ) : null}
        </div>
      )}
    </div>
  );
}
