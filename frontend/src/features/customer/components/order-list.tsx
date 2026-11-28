"use client";

import Link from "next/link";
import { useState } from "react";

import { StorefrontEmptyState, StorefrontErrorState } from "@/components/layouts/storefront-state";
import { AsyncPanel } from "@/components/ui/async-panel";
import { Button } from "@/components/ui/button";
import { orderStatusToPillVariant, StatusPill } from "@/components/ui/status-pill";
import { ROUTE } from "@/constants/routes";
import { CatalogPagination } from "@/features/catalog";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPickupDateTime } from "@/lib/validation/pickup";
import { plural } from "@/lib/utils";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useOrders } from "../hooks";
import { orderHeadline } from "../lib/order-progress";
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
    return (
      <div aria-busy="true" className="storefront-orders__list">
        {[0, 1, 2].map((row) => (
          <span className="skeleton block h-16 w-full" key={row} />
        ))}
      </div>
    );
  }

  // Kept while a filtered request fails, so the filter that caused it can be changed.
  const orders = ordersQuery.data?.orders ?? [];
  const pagination = ordersQuery.data?.pagination;

  if (!ordersQuery.isError && orders.length === 0 && !filtered && !refetching) {
    return (
      <StorefrontEmptyState
        action={
          <Button asChild>
            <Link href={ROUTE.products}>Browse all products</Link>
          </Button>
        }
        message="Your orders and their pickups appear here once you place one."
        title="No orders yet"
      />
    );
  }

  return (
    <div className="storefront-orders">
      <OrderHistoryFilters value={filters} onChange={changeFilters} />

      <AsyncPanel refetching={refetching}>
        {ordersQuery.isError ? (
          <StorefrontErrorState onRetry={() => void ordersQuery.refetch()} title="We could not load your orders" />
        ) : orders.length === 0 ? (
          <StorefrontEmptyState
            action={
              <Button onClick={() => changeFilters(NO_FILTERS)} size="sm" type="button" variant="outline">
                Clear filters
              </Button>
            }
            title="No orders match these filters"
          />
        ) : (
          <ul className="storefront-orders__list">
            {orders.map((order) => (
              <li key={order.id}>
                <Link className="storefront-order-row" href={ROUTE.orderDetail(order.id)}>
                  <span className="storefront-order-row__id">
                    <span className="text-order-code">{order.code}</span>
                    <span className="storefront-line-meta">
                      Placed {formatDateTime(order.created_at)} · {plural(order.item_count, "item")}
                    </span>
                    {order.unread_messages > 0 ? (
                      <span className="storefront-order-row__unread">
                        {plural(order.unread_messages, "new message")}
                      </span>
                    ) : null}
                  </span>
                  <span className="storefront-order-row__pickup">
                    {order.pickup_at ? `Pickup ${formatPickupDateTime(order.pickup_at)}` : "Pickup at the counter"}
                  </span>
                  <span className="storefront-order-row__total text-price">{formatPriceCents(order.total_cents)}</span>
                  <span className="storefront-order-row__status">
                    <StatusPill label={orderHeadline(order.status)} variant={orderStatusToPillVariant(order.status)} />
                  </span>
                  <span aria-hidden="true" className="storefront-order-row__go">
                    →
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </AsyncPanel>

      {pagination ? (
        <CatalogPagination onPageChange={setPage} page={page} totalPages={pagination.total_pages} />
      ) : null}
    </div>
  );
}
