"use client";

import { useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import type { Fulfillment } from "@/lib/schemas/order";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useCancelOrder } from "../hooks";
import type { Order } from "../schemas";
import { OrderReschedule } from "./order-reschedule";

const RESCHEDULE_ID = "order-reschedule";

type OrderChangesProps = {
  order: Order;
  /** What the order's items ask of the bakery, for choosing another pickup. */
  fulfillment: Fulfillment;
};

/** Moving the pickup or cancelling, open until the bakery starts making the order. */
export function OrderChanges({ order, fulfillment }: OrderChangesProps) {
  const [moving, setMoving] = useState(false);
  const [cancelling, setCancelling] = useState(false);
  const changeButton = useRef<HTMLButtonElement>(null);
  const cancel = useCancelOrder(order.id);
  const paid = order.payment?.captured_at != null;
  // Whether the money is taken is PayPal's to decide first.
  const inReview = order.payment?.status === "pending";

  function closePicker(): void {
    setMoving(false);
    changeButton.current?.focus();
  }

  return (
    <div className="storefront-order-changes">
      <p className="text-caption">You can change the pickup time or cancel until we start making your order.</p>
      <div className="storefront-order-changes__actions">
        <Button
          ref={changeButton}
          aria-controls={RESCHEDULE_ID}
          aria-expanded={moving}
          disabled={cancel.isPending}
          size="sm"
          type="button"
          variant="outline"
          onClick={() => setMoving(!moving)}
        >
          Change pickup time
        </Button>
        <button
          className="storefront-text-action storefront-text-action--danger"
          disabled={cancel.isPending || inReview}
          title={inReview ? "PayPal is still reviewing your payment" : undefined}
          type="button"
          onClick={() => setCancelling(true)}
        >
          Cancel order
        </button>
      </div>
      {moving ? (
        <OrderReschedule fulfillment={fulfillment} id={RESCHEDULE_ID} orderId={order.id} onDone={closePicker} />
      ) : null}
      <ConfirmDialog
        cancelLabel="Keep order"
        confirmLabel="Cancel order"
        confirmVariant="destructive"
        description={
          paid
            ? `We refund ${formatPriceCents(order.total_cents)} to your PayPal account, and a discount code you used can be used again.`
            : "Your pickup time is released, and a discount code you used can be used again."
        }
        isPending={cancel.isPending}
        open={cancelling}
        title={`Cancel order ${order.code}?`}
        onCancel={() => setCancelling(false)}
        onConfirm={() => cancel.mutate(undefined, { onSettled: () => setCancelling(false) })}
      />
    </div>
  );
}
