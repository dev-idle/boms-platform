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
    return <p className="text-sm text-zinc-500">Loading orders…</p>;
  }

  if (ordersQuery.isError) {
    return <p className="text-sm text-red-600">Failed to load orders.</p>;
  }

  const orders = ordersQuery.data?.orders ?? [];
  const pagination = ordersQuery.data?.pagination;

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-2">
        {STATUS_FILTERS.map((option) => (
          <button
            key={option.label}
            className={`rounded-full px-4 py-2 text-sm font-medium ${
              status === option.value
                ? "bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900"
                : "bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200"
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
        <p className="text-sm text-zinc-600 dark:text-zinc-400">
          No orders match this filter.
        </p>
      ) : (
        <ul className="divide-y divide-zinc-200 rounded-lg border border-zinc-200 dark:divide-zinc-800 dark:border-zinc-800">
          {orders.map((order) => (
            <li
              key={order.id}
              className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p className="font-medium text-zinc-900 dark:text-zinc-50">
                  {formatPriceCents(order.total_cents)} · {order.status}
                </p>
                <p className="text-sm text-zinc-500">
                  {formatDateTime(order.created_at)} · {order.item_count} item
                  {order.item_count === 1 ? "" : "s"}
                </p>
                <p className="text-sm text-zinc-600 dark:text-zinc-400">
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
          <span className="text-sm text-zinc-600">
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
