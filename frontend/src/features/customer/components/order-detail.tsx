"use client";

import Link from "next/link";
import { z } from "zod";

import { StorefrontPageHeader } from "@/components/layouts/storefront-page-header";
import { StorefrontEmptyState, StorefrontErrorState } from "@/components/layouts/storefront-state";
import { Button } from "@/components/ui/button";
import { CustomizationSummary } from "@/components/ui/customization-summary";
import { StatusPill, ticketStatusToPillVariant } from "@/components/ui/status-pill";
import { ROUTE } from "@/constants/routes";
import { CatalogTile } from "@/features/catalog";
import { isApiError } from "@/lib/errors";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
import { formatOrderTypeLabel } from "@/lib/schemas/order";
import { STATION_LABEL, type TicketStatus } from "@/lib/schemas/ticket";
import { formatPriceCents } from "@/lib/validation/catalog";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPickupDateTime } from "@/lib/validation/pickup";

import { useOrder, usePayPalReturn } from "../hooks";
import { isWithTheBakery, orderHeadline } from "../lib/order-progress";
import { reviewableProducts } from "../lib/reviewable-products";
import { OrderChanges } from "./order-changes";
import { OrderMessages } from "./order-messages";
import { OrderPayment } from "./order-payment";
import { OrderProgressStepper } from "./order-progress-stepper";
import { OrderReviews } from "./order-reviews";

type OrderDetailProps = {
  orderId: string;
};

const ORDERS_BACK = { href: ROUTE.orders, label: PAGE_TITLES.orders };

/** The title while there is no order to name. */
const ORDER_TITLE = "Order";

/** How a station's part of the order reads to the customer. */
const TICKET_STATUS_LABEL: Record<TicketStatus, string> = {
  queued: "Queued",
  in_progress: "In progress",
  ready: "Ready",
  cancelled: "Cancelled",
};

/** An order that is not the customer's, or no longer exists. */
function OrderMissing() {
  return (
    <div className="storefront-page-stack">
      <StorefrontPageHeader back={ORDERS_BACK} title={ORDER_TITLE} />
      <StorefrontEmptyState
        action={
          <Button asChild variant="outline">
            <Link href={ROUTE.orders}>See your orders</Link>
          </Button>
        }
        message="The link may be mistyped, or the order belongs to another account."
        title="We could not find this order"
      />
    </div>
  );
}

/** The page's own shape while the order loads: its header, the work, the rail. */
function OrderDetailSkeleton() {
  return (
    <div aria-busy="true" className="storefront-page-stack">
      <div className="grid gap-5">
        <span className="skeleton block h-5 w-24" />
        <span className="skeleton block h-12 w-72 max-w-full" />
      </div>
      <div className="storefront-detail-layout">
        <div className="grid content-start gap-6">
          <span className="skeleton block h-10 w-full" />
          <span className="skeleton block h-24 w-full" />
        </div>
        <span className="skeleton block h-72 w-full" />
      </div>
    </div>
  );
}

export function OrderDetail({ orderId }: OrderDetailProps) {
  const isValidId = z.uuid().safeParse(orderId).success;
  const orderQuery = useOrder(orderId);
  const paying = usePayPalReturn(orderId);

  if (!isValidId) {
    return <OrderMissing />;
  }

  if (orderQuery.isPending) {
    return <OrderDetailSkeleton />;
  }

  if (orderQuery.isError) {
    if (isApiError(orderQuery.error) && orderQuery.error.status === 404) {
      return <OrderMissing />;
    }
    return (
      <div className="storefront-page-stack">
        <StorefrontPageHeader back={ORDERS_BACK} title={ORDER_TITLE} />
        <StorefrontErrorState
          message="Check your connection and try again."
          onRetry={() => void orderQuery.refetch()}
          title="We could not load this order"
        />
      </div>
    );
  }

  const order = orderQuery.data;
  const reviewable = reviewableProducts(order);
  const showStations = isWithTheBakery(order.status) && order.tickets.length > 0;

  return (
    <div className="storefront-page-stack">
      <StorefrontPageHeader
        back={ORDERS_BACK}
        lead={
          <>
            <span className="text-order-code">{order.code}</span> · Placed {formatDateTime(order.created_at)} ·{" "}
            {formatOrderTypeLabel(order.order_type)}
          </>
        }
        title={orderHeadline(order.status)}
      />

      <div className="storefront-detail-layout">
        <div className="storefront-order-main">
          <OrderProgressStepper placedAt={order.created_at} status={order.status} timeline={order.timeline} />

          {order.status === "pending" ? (
            <p className="storefront-order-note">
              We are checking your reference photo and will email you once we accept your order, within 24 hours of
              your payment. If we cannot make it, or have not accepted it by then, you are refunded in full.
            </p>
          ) : null}

          <section aria-labelledby="order-items" className="storefront-order-section">
            <h2 className="storefront-section-heading" id="order-items">
              Items
            </h2>
            <ul className="storefront-order-lines">
              {order.items.map((item) => (
                <li className="storefront-order-line" key={item.id}>
                  <CatalogTile name={item.name} scale="row" url={item.image_url} />
                  <div className="storefront-order-line__copy">
                    <p className="storefront-line-name">{item.name}</p>
                    <p className="storefront-line-meta">
                      {item.quantity} × <span className="text-price">{formatPriceCents(item.unit_price_cents)}</span>
                    </p>
                    {item.customization ? <CustomizationSummary customization={item.customization} /> : null}
                  </div>
                  <p className="storefront-order-line__total text-price">{formatPriceCents(item.line_total_cents)}</p>
                </li>
              ))}
            </ul>
          </section>

          {reviewable.length > 0 ? <OrderReviews orderId={order.id} products={reviewable} /> : null}
          {order.can_message ? <OrderMessages orderId={order.id} /> : null}
        </div>

        <aside aria-label="Pickup and payment" className="storefront-order-rail storefront-detail-rail">
          {order.status === "awaiting_payment" ? <OrderPayment confirming={paying.isPending} order={order} /> : null}

          <section aria-labelledby="order-pickup" className="storefront-order-rail__group">
            <h2 className="storefront-order-rail__label" id="order-pickup">
              Pickup
            </h2>
            {order.pickup_at ? (
              <p className="storefront-order-rail__value">{formatPickupDateTime(order.pickup_at)}</p>
            ) : (
              <p className="storefront-order-rail__text">At the counter</p>
            )}
            {order.pickup_code ? (
              <div className="storefront-order-rail__code">
                <p className="storefront-order-rail__digits">{order.pickup_code}</p>
                <p className="text-caption">Give this code at the counter when you pick up your order.</p>
              </div>
            ) : null}
            {/* The API sends what the items need only while the order may still change. */}
            {order.fulfillment ? <OrderChanges fulfillment={order.fulfillment} order={order} /> : null}
          </section>

          {showStations ? (
            <section aria-labelledby="order-preparation" className="storefront-order-rail__group">
              <h2 className="storefront-order-rail__label" id="order-preparation">
                Preparation
              </h2>
              <ul className="storefront-order-rail__stations">
                {order.tickets.map((ticket) => (
                  <li key={ticket.station}>
                    <span>{STATION_LABEL[ticket.station]}</span>
                    <StatusPill
                      label={TICKET_STATUS_LABEL[ticket.status]}
                      variant={ticketStatusToPillVariant(ticket.status)}
                    />
                  </li>
                ))}
              </ul>
            </section>
          ) : null}

          <div className="storefront-order-rail__group">
            <dl className="storefront-cart-summary__rows storefront-order-rail__rows">
              <div className="storefront-cart-summary__row">
                <dt>Subtotal</dt>
                <dd className="text-price">{formatPriceCents(order.subtotal_cents)}</dd>
              </div>
              {order.discount_cents > 0 ? (
                <div className="storefront-cart-summary__row">
                  <dt>
                    Discount
                    {order.discount_code_snapshot ? (
                      <>
                        {" "}
                        <span className="storefront-cart-summary__code">{order.discount_code_snapshot}</span>
                      </>
                    ) : null}
                  </dt>
                  <dd className="text-price">−{formatPriceCents(order.discount_cents)}</dd>
                </div>
              ) : null}
            </dl>
            <div className="storefront-cart-summary__total">
              <span className="storefront-cart-summary__total-label">Total</span>
              <span className="text-price">{formatPriceCents(order.total_cents)}</span>
            </div>
            {order.payment?.provider === "cash" ? (
              order.payment.captured_at ? (
                <p className="text-caption">Paid in cash on {formatDateTime(order.payment.captured_at)}</p>
              ) : isWithTheBakery(order.status) ? (
                <p className="text-caption">You pay in cash when you pick it up.</p>
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
        </aside>
      </div>
    </div>
  );
}
