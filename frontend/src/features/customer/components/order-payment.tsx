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
    <section aria-labelledby="order-payment" className="storefront-order-detail__payment">
      <h2 className="text-form-label" id="order-payment">
        Payment
      </h2>
      {confirming ? (
        <p className="text-caption">Confirming your payment…</p>
      ) : order.payment?.status === "pending" ? (
        <p className="text-caption">
          PayPal is reviewing your payment. Your order is confirmed as soon as it clears.
        </p>
      ) : (
        <>
          <p className="text-caption">
            Pay {formatPriceCents(order.total_cents)} with PayPal
            {order.payment_due_at ? (
              <>
                {" by "}
                <time dateTime={order.payment_due_at}>{formatDateTime(order.payment_due_at)}</time>
              </>
            ) : null}{" "}
            to confirm your order. We hold your pickup time until then.
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
        </>
      )}
    </section>
  );
}
