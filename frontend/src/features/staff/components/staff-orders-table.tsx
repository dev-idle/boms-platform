"use client";

import Link from "next/link";
import { useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
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
      <div className="flex flex-wrap gap-2">
        {STATUS_FILTERS.map((option) => (
          <button
            key={option.label}
            className={`min-h-11 rounded-full px-4 py-2 text-sm font-medium transition-colors duration-standard ease-default ${
              status === option.value
                ? "bg-rose-500 text-surface"
                : "bg-blush text-ink-2 hover:text-rose-500"
            }`}
            onClick={() => {
              setStatus(option.value);
              setPage(1);
            }}
            type="button"
          >
            {option.label}
          </button>
        ))}
      </div>

      {orders.length === 0 ? (
        <p className="text-sm text-ink-2">
          No orders match this filter.
        </p>
      ) : (
        <ul className="divide-y divide-border rounded-card border border-border">
          {orders.map((order) => (
            <li
              key={order.id}
              className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p className="font-medium text-ink">
                  {formatPriceCents(order.total_cents)} · {order.status}
                </p>
                <p className="text-sm text-muted">
                  {formatDateTime(order.created_at)} · {order.item_count} item
                  {order.item_count === 1 ? "" : "s"}
                </p>
                <p className="text-sm text-ink-2">
                  {order.customer.display_name
                    ? `${order.customer.display_name} · `
                    : ""}
                  {order.customer.email}
                </p>
              </div>
              <Link href={ROUTE.staff.orderDetail(order.id)}>
                <Button type="button" variant="outline">
                  View
                </Button>
              </Link>
            </li>
          ))}
        </ul>
      )}

      {pagination && pagination.total_pages > 1 ? (
        <div className="flex items-center gap-2">
          <Button
            disabled={page <= 1}
            onClick={() => setPage((current) => Math.max(1, current - 1))}
            type="button"
            variant="outline"
          >
            Previous
          </Button>
          <span className="text-sm text-ink-2">
            Page {pagination.page} of {pagination.total_pages}
          </span>
          <Button
            disabled={page >= pagination.total_pages}
            onClick={() => setPage((current) => current + 1)}
            type="button"
            variant="outline"
          >
            Next
          </Button>
        </div>
      ) : null}
    </div>
  );
}
