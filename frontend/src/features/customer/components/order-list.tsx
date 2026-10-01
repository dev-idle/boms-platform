"use client";

import Link from "next/link";
import { useState } from "react";

import { AsyncPanel } from "@/components/ui/async-panel";
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
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPickupDateTime } from "@/lib/validation/pickup";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useOrders } from "../hooks";
import { OrderHistoryFilters, type OrderHistoryFilterValue } from "./order-history-filters";

const NO_FILTERS: OrderHistoryFilterValue = { status: undefined, from: "", to: "" };

export function OrderList() {
  const [page, setPage] = useState(1);
  const [filters, setFilters] = useState(NO_FILTERS);
  const ordersQuery = useOrders({
    page,
    page_size: 20,
    status: filters.status,
    from: filters.from || undefined,
    to: filters.to || undefined,
  });
  const { refetching } = getQuerySurface(ordersQuery);
  const filtered = filters.status !== undefined || filters.from !== "" || filters.to !== "";

  function changeFilters(next: OrderHistoryFilterValue): void {
    setFilters(next);
    setPage(1);
  }

  if (ordersQuery.isPending) {
    return <InlineLoadingState />;
  }

  if (ordersQuery.isError && isApiError(ordersQuery.error) && ordersQuery.error.isAuthError()) {
    return <p className="text-sm text-error">Sign in to view your orders.</p>;
  }

  // Kept while a filtered request fails, so the filter that caused it can be changed.
  const orders = ordersQuery.data?.orders ?? [];
  const pagination = ordersQuery.data?.pagination;

  if (!ordersQuery.isError && orders.length === 0 && !filtered && !refetching) {
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
      <OrderHistoryFilters value={filters} onChange={changeFilters} />

      <AsyncPanel refetching={refetching}>
        {ordersQuery.isError ? (
          <div className="storefront-empty-state">
            <p className="storefront-empty-state__message text-error">Failed to load orders.</p>
            <Button type="button" variant="outline" onClick={() => void ordersQuery.refetch()}>
              Try again
            </Button>
          </div>
        ) : orders.length === 0 ? (
          <div className="storefront-empty-state">
            <p className="storefront-empty-state__message">No orders match these filters.</p>
            <Button type="button" variant="outline" onClick={() => changeFilters(NO_FILTERS)}>
              Clear filters
            </Button>
          </div>
        ) : (
          <ul className="storefront-orders__list">
            {orders.map((order) => (
              <li key={order.id}>
                <Link className="storefront-order-card" href={ROUTE.orderDetail(order.id)}>
                  <div className="storefront-order-card__main">
                    <p className="text-order-code">{order.code}</p>
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
                      {order.pickup_at
                        ? ` · Pickup ${formatPickupDateTime(order.pickup_at)}`
                        : ""}
                    </p>
                    {order.unread_messages > 0 ? (
                      <p className="storefront-order-card__unread">
                        {order.unread_messages} new {order.unread_messages === 1 ? "message" : "messages"}
                      </p>
                    ) : null}
                  </div>
                  <span className="storefront-order-card__cta" aria-hidden="true">
                    →
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </AsyncPanel>

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
