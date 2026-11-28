"use client";

import { Button } from "@/components/ui/button";
import { formatPriceCents } from "@/lib/validation/catalog";
import { formatDateTime } from "@/lib/validation/datetime";

import { useStartPayment } from "../hooks";
import type { Order } from "../schemas";

type OrderPaymentProps = {
  order: Order;
  /** The payment the customer just approved is being taken. */
  confirming: boolean;
};

/**
 * Pays for an order placed online: the customer approves on PayPal's own page
 * and comes back here, where the payment is taken.
 */
export function OrderPayment({ order, confirming }: OrderPaymentProps) {
  const start = useStartPayment(order.id);

  return (
    <section aria-labelledby="order-payment" className="storefront-order-rail__group">
      <h2 className="storefront-order-rail__label" id="order-payment">
        Payment due
      </h2>
      {confirming ? (
        <p className="text-caption">Confirming your payment…</p>
      ) : order.payment?.status === "pending" ? (
        <p className="text-caption">
          PayPal is reviewing your payment. We will update your order as soon as it clears.
        </p>
      ) : (
        <>
          <p className="storefront-order-rail__text">
            Pay <span className="text-price">{formatPriceCents(order.total_cents)}</span> with PayPal
            {order.payment_due_at ? (
              <>
                {" by "}
                <time dateTime={order.payment_due_at}>{formatDateTime(order.payment_due_at)}</time>
              </>
            ) : null}{" "}
            to keep your pickup time.
          </p>
          <Button
            aria-busy={start.isPending || undefined}
            className="w-full"
            disabled={start.isPending}
            onClick={() => start.mutate()}
            type="button"
          >
            {start.isPending ? "Opening PayPal…" : "Pay with PayPal"}
          </Button>
          <p className="text-caption">
            You pay on PayPal&apos;s secure page. We never see or store your card or bank details.
          </p>
        </>
      )}
    </section>
  );
}
