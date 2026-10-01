"use client";

import { Button } from "@/components/ui/button";

import { useRescheduleOrder } from "../hooks";
import { usePickupChoice } from "../hooks/use-pickup-choice";
import type { Fulfillment } from "../schemas";
import { PickupSlotPicker } from "./pickup-slot-picker";

const ERROR_ID = "reschedule-error";

type OrderRescheduleProps = {
  id: string;
  orderId: string;
  fulfillment: Fulfillment;
  /** Closes the picker: the pickup moved, or the customer keeps it. */
  onDone: () => void;
};

/** Another pickup day and slot for an order the bakery has not started, under the checkout rules. */
export function OrderReschedule({ id, orderId, fulfillment, onDone }: OrderRescheduleProps) {
  const reschedule = useRescheduleOrder(orderId);
  const choice = usePickupChoice(fulfillment, reschedule.isPending);

  function move(): void {
    const pickupAt = choice.ready ? choice.pickupAt() : null;
    if (pickupAt === null) {
      return;
    }
    reschedule.mutate({ pickup_at: pickupAt }, { onSuccess: onDone });
  }

  return (
    <div className="storefront-order-detail__reschedule" id={id}>
      <div>
        <PickupSlotPicker {...choice.picker} errorId={ERROR_ID} />
        <p className="storefront-pickup-picker__error text-caption" id={ERROR_ID} role="status">
          {choice.message}
        </p>
      </div>
      <div className="storefront-order-detail__change-actions">
        <Button
          aria-busy={reschedule.isPending || undefined}
          disabled={!choice.ready || reschedule.isPending}
          type="button"
          onClick={move}
        >
          {reschedule.isPending ? "Moving pickup…" : "Move pickup"}
        </Button>
        <Button disabled={reschedule.isPending} type="button" variant="outline" onClick={onDone}>
          Keep time
        </Button>
        {choice.retry ? (
          <Button type="button" variant="outline" onClick={choice.retry}>
            Try again
          </Button>
        ) : null}
      </div>
    </div>
  );
}
