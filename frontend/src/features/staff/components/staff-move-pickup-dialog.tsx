"use client";

import { AppDialog, AppDialogFooterActions } from "@/components/ui/app-dialog";
import { Button } from "@/components/ui/button";
import { PickupSlotPicker, usePickupChoice } from "@/features/customer";

import { useMoveStaffOrderPickup } from "../hooks";
import { NO_ITEM_NEEDS } from "../lib/pickup-schedule";

const ERROR_ID = "staff-move-pickup-error";

type StaffMovePickupDialogProps = {
  orderId: string;
  orderCode: string;
  onClose: () => void;
};

/**
 * Another pickup for an order not handed over yet: a free slot ahead within the
 * bakery's days and hours. Whether the kitchen still has its notice is the
 * counter's call, so no slot is held back for it.
 */
export function StaffMovePickupDialog({ orderId, orderCode, onClose }: StaffMovePickupDialogProps) {
  const move = useMoveStaffOrderPickup(orderId);
  const choice = usePickupChoice(NO_ITEM_NEEDS, move.isPending);

  function submit(): void {
    const pickupAt = choice.ready ? choice.pickupAt() : null;
    if (pickupAt === null) {
      return;
    }
    move.mutate({ pickup_at: pickupAt }, { onSuccess: onClose });
  }

  return (
    <AppDialog
      description={`The customer sees the new time for ${orderCode} on their order page.`}
      footer={
        <AppDialogFooterActions>
          <Button disabled={move.isPending} type="button" variant="outline" onClick={onClose}>
            Keep time
          </Button>
          <Button
            aria-busy={move.isPending || undefined}
            disabled={!choice.ready || move.isPending}
            type="button"
            onClick={submit}
          >
            {move.isPending ? "Moving pickup…" : "Move pickup"}
          </Button>
        </AppDialogFooterActions>
      }
      isPending={move.isPending}
      open
      title="Move this pickup?"
      onClose={onClose}
    >
      <div className="staff-move-pickup">
        <PickupSlotPicker {...choice.picker} errorId={ERROR_ID} />
        <p className="storefront-pickup-picker__error text-caption" id={ERROR_ID} role="status">
          {choice.message}
        </p>
        {choice.retry ? (
          <Button className="justify-self-start" type="button" variant="outline" onClick={choice.retry}>
            Try again
          </Button>
        ) : null}
      </div>
    </AppDialog>
  );
}
