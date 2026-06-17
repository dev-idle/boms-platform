"use client";

import { z } from "zod";

import { InlineLoadingState } from "@/components/ui/loading-state";
import {
  formatOrderStatusLabel,
  orderStatusToPillVariant,
  StatusPill,
} from "@/components/ui/status-pill";
import { isApiError } from "@/lib/errors";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useOrder } from "../hooks";

type OrderDetailProps = {
  orderId: string;
};

export function OrderDetail({ orderId }: OrderDetailProps) {
  const isValidId = z.string().uuid().safeParse(orderId).success;
  const orderQuery = useOrder(orderId);

  if (!isValidId) {
    return <p className="text-sm text-muted">Invalid order link.</p>;
  }

  if (orderQuery.isPending) {
    return <InlineLoadingState className="storefront-customer-loading" />;
  }

  if (orderQuery.isError) {
    const message =
      isApiError(orderQuery.error) && orderQuery.error.status === 404
        ? "Order not found."
        : "Failed to load order.";
    return <p className="text-sm text-error">{message}</p>;
  }

  const order = orderQuery.data;
  if (!order) {
    return <p className="text-sm text-muted">Order not found.</p>;
  }

  return (
    <div className="storefront-order-detail">
      <div className="storefront-panel storefront-order-detail__card">
        <div className="storefront-order-detail__meta">
          <p className="text-caption">
            Placed {formatDateTime(order.created_at)}
          </p>
          <StatusPill
            label={formatOrderStatusLabel(order.status)}
            variant={orderStatusToPillVariant(order.status)}
          />
        </div>

        {order.discount_code_snapshot ? (
          <p className="text-caption">
            Discount code: {order.discount_code_snapshot}
          </p>
        ) : null}

        <ul className="storefront-order-detail__items">
          {order.items.map((item) => (
            <li key={item.id} className="storefront-order-detail__item">
              <div>
                <p className="text-section-heading">
                  {item.quantity}× {item.name}
                </p>
                <p className="text-caption">
                  {formatPriceCents(item.unit_price_cents)} each
                </p>
              </div>
              <p className="text-table-cell">
                {formatPriceCents(item.line_total_cents)}
              </p>
            </li>
          ))}
        </ul>

        <div className="storefront-cart-summary storefront-order-detail__totals">
          <div className="storefront-cart-summary__row">
            <span className="text-muted">Subtotal</span>
            <span className="text-table-cell">
              {formatPriceCents(order.subtotal_cents)}
            </span>
          </div>
          {order.discount_cents > 0 ? (
            <div className="storefront-cart-summary__row storefront-cart-summary__row--discount">
              <span>Discount</span>
              <span>-{formatPriceCents(order.discount_cents)}</span>
            </div>
          ) : null}
          <div className="storefront-cart-summary__row storefront-cart-summary__row--total">
            <span>Total</span>
            <span className="text-price">
              {formatPriceCents(order.total_cents)}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
