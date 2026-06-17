"use client";

import Link from "next/link";
import { useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import { InlineLoadingState } from "@/components/ui/loading-state";
import {
  formatOrderStatusLabel,
  orderStatusToPillVariant,
  StatusPill,
} from "@/components/ui/status-pill";
import { ROUTE } from "@/constants/routes";
import { STOREFRONT_NAV_COPY } from "@/constants/storefront-nav-copy";
import { isApiError } from "@/lib/errors";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useOrders } from "../hooks";

export function OrderList() {
  const [page, setPage] = useState(1);
  const filter = useMemo(() => ({ page, page_size: 20 }), [page]);
  const ordersQuery = useOrders(filter);

  if (ordersQuery.isPending) {
    return <InlineLoadingState className="storefront-customer-loading" />;
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
      <div className="storefront-empty-state storefront-empty-state--card">
        <p className="storefront-empty-state__message">
          You have not placed any orders yet.
        </p>
        <Button asChild variant="outline">
          <Link href={ROUTE.products}>{STOREFRONT_NAV_COPY.returnToShop}</Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="storefront-orders">
      <ul className="storefront-orders__list">
        {orders.map((order) => (
          <li key={order.id}>
            <Link
              className="storefront-order-card"
              href={ROUTE.orderDetail(order.id)}
            >
              <div className="storefront-order-card__main">
                <div className="storefront-order-card__top">
                  <p className="storefront-order-card__price text-price">
                    {formatPriceCents(order.total_cents)}
                  </p>
                  <StatusPill
                    label={formatOrderStatusLabel(order.status)}
                    variant={orderStatusToPillVariant(order.status)}
                  />
                </div>
                <p className="storefront-order-card__meta text-caption">
                  {formatDateTime(order.created_at)} · {order.item_count} item
                  {order.item_count === 1 ? "" : "s"}
                </p>
              </div>
              <span className="storefront-order-card__cta" aria-hidden="true">
                →
              </span>
            </Link>
          </li>
        ))}
      </ul>

      {pagination && pagination.total_pages > 1 ? (
        <div className="storefront-orders__pagination">
          <Button
            disabled={page <= 1}
            type="button"
            variant="outline"
            onClick={() => setPage((current) => Math.max(1, current - 1))}
          >
            Previous
          </Button>
          <span className="text-caption">
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
