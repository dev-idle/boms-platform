"use client";

import { z } from "zod";

import { CustomizationSummary } from "@/components/ui/customization-summary";
import { InlineLoadingState } from "@/components/ui/loading-state";
import {
  formatOrderStatusLabel,
  orderStatusToPillVariant,
  StatusPill,
  ticketStatusToPillVariant,
} from "@/components/ui/status-pill";
import { isApiError } from "@/lib/errors";
import { formatOrderTypeLabel } from "@/lib/schemas/order";
import { STATION_LABEL } from "@/lib/schemas/ticket";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPickupDateTime } from "@/lib/validation/pickup";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useOrder, usePayPalReturn } from "../hooks";
import { isWithTheBakery } from "../lib/order-progress";
import { OrderChanges } from "./order-changes";
import { OrderPayment } from "./order-payment";
import { OrderProgressStepper } from "./order-progress-stepper";

type OrderDetailProps = {
  orderId: string;
};

export function OrderDetail({ orderId }: OrderDetailProps) {
  const isValidId = z.uuid().safeParse(orderId).success;
  const orderQuery = useOrder(orderId);
  const paying = usePayPalReturn(orderId);

  if (!isValidId) {
    return <p className="text-sm text-muted">Invalid order link.</p>;
  }

  if (orderQuery.isPending) {
    return <InlineLoadingState />;
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
        <p className="text-order-code">{order.code}</p>
        <div className="storefront-order-detail__meta">
          <div className="storefront-order-detail__facts">
            <p className="text-caption">
              Placed {formatDateTime(order.created_at)}
            </p>
            {order.pickup_at ? (
              <p className="text-caption">
                Pickup {formatPickupDateTime(order.pickup_at)}
              </p>
            ) : null}
            <p className="text-caption">{formatOrderTypeLabel(order.order_type)}</p>
          </div>
          <StatusPill
            label={formatOrderStatusLabel(order.status)}
            variant={orderStatusToPillVariant(order.status)}
          />
        </div>

        <OrderProgressStepper status={order.status} timeline={order.timeline} />

        {order.pickup_code ? (
          <section aria-labelledby="order-pickup-code" className="storefront-order-detail__pickup-code">
            <h2 className="text-form-label" id="order-pickup-code">
              Pickup code
            </h2>
            <p className="storefront-order-detail__pickup-digits">{order.pickup_code}</p>
            <p className="text-caption">Give this code at the counter when you collect your order.</p>
          </section>
        ) : null}

        {order.status === "pending" ? (
          <p className="text-caption">
            We are checking your custom order and will email you once we accept it. If we cannot make it, you are
            refunded in full.
          </p>
        ) : null}

        {order.status === "awaiting_payment" ? <OrderPayment confirming={paying.isPending} order={order} /> : null}

        {/* The API sends what the items need only while the order may still change. */}
        {order.fulfillment ? (
          <OrderChanges fulfillment={order.fulfillment} order={order} />
        ) : null}

        {isWithTheBakery(order.status) && order.tickets.length > 0 ? (
          <section
            aria-labelledby="order-preparation"
            className="storefront-order-detail__preparation"
          >
            <h2 className="text-form-label" id="order-preparation">
              Preparation
            </h2>
            <ul className="storefront-order-detail__stations">
              {order.tickets.map((ticket) => (
                <li key={ticket.station} className="storefront-order-detail__station">
                  <span className="text-caption">{STATION_LABEL[ticket.station]}</span>
                  <StatusPill
                    label={formatOrderStatusLabel(ticket.status)}
                    variant={ticketStatusToPillVariant(ticket.status)}
                  />
                </li>
              ))}
            </ul>
          </section>
        ) : null}

        {order.discount_code_snapshot ? (
          <p className="text-caption">
            Discount code: {order.discount_code_snapshot}
          </p>
        ) : null}

        <ul className="storefront-order-detail__items">
          {order.items.map((item) => (
            <li key={item.id} className="storefront-order-detail__item">
              <div className="grid gap-1">
                <p className="text-section-heading">
                  {item.quantity}× {item.name}
                </p>
                <p className="text-caption">
                  {formatPriceCents(item.unit_price_cents)} each
                </p>
                {item.customization ? <CustomizationSummary customization={item.customization} /> : null}
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
          {order.payment?.provider === "cash" ? (
            order.payment.captured_at ? (
              <p className="text-caption">Paid in cash on {formatDateTime(order.payment.captured_at)}</p>
            ) : isWithTheBakery(order.status) ? (
              <p className="text-caption">You pay in cash when you collect it.</p>
            ) : null
          ) : order.payment?.captured_at ? (
            <p className="text-caption">Paid with PayPal on {formatDateTime(order.payment.captured_at)}</p>
          ) : null}
          {order.payment?.refunded_at ? (
            <p className="text-caption">Refunded to PayPal on {formatDateTime(order.payment.refunded_at)}</p>
          ) : order.payment?.refund_requested_at ? (
            <p className="text-caption">Refund to PayPal in progress</p>
          ) : null}
        </div>
      </div>
    </div>
  );
}
