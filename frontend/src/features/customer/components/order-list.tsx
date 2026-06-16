"use client";

import Link from "next/link";
import { useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useOrders } from "../hooks";

export function OrderList() {
  const [page, setPage] = useState(1);
  const filter = useMemo(() => ({ page, page_size: 20 }), [page]);
  const ordersQuery = useOrders(filter);

  if (ordersQuery.isPending) {
    return <InlineLoadingState />;
  }

  if (ordersQuery.isError) {
    const message =
      isApiError(ordersQuery.error) && ordersQuery.error.isAuthError()
        ? "Sign in to view your orders."
        : "Failed to load orders.";
    return <p className="text-sm text-error">{message}</p>;
  }

  const orders = ordersQuery.data?.orders ?? [];
  const pagination = ordersQuery.data?.pagination;

  if (orders.length === 0) {
    return (
      <p className="text-sm text-muted">
        You have not placed any orders yet.
      </p>
    );
  }

  return (
    <div className="space-y-4">
      <ul className="divide-y divide-border rounded-card border border-border">
        {orders.map((order) => (
          <li key={order.id} className="flex items-center justify-between gap-4 p-4">
            <div>
              <p className="font-medium text-ink">
                {formatPriceCents(order.total_cents)}
              </p>
              <p className="text-sm text-muted">
                {formatDateTime(order.created_at)} · {order.item_count} item
                {order.item_count === 1 ? "" : "s"} · {order.status}
              </p>
            </div>
            <Link href={ROUTE.orderDetail(order.id)}>
              <Button type="button" variant="outline">
                View
              </Button>
            </Link>
          </li>
        ))}
      </ul>

      {pagination && pagination.total_pages > 1 ? (
        <div className="flex items-center gap-2">
          <Button
            disabled={page <= 1}
            type="button"
            variant="outline"
            onClick={() => setPage((current) => Math.max(1, current - 1))}
          >
            Previous
          </Button>
          <span className="text-sm text-muted">
            Page {pagination.page} of {pagination.total_pages}
          </span>
          <Button
            disabled={page >= pagination.total_pages}
            type="button"
            variant="outline"
            onClick={() => setPage((current) => current + 1)}
          >
            Next
          </Button>
        </div>
      ) : null}
    </div>
  );
}
