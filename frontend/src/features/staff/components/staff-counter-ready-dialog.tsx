"use client";

import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { formatTicketItems } from "@/lib/schemas/ticket";

import { usePatchCounterTicketStatus } from "../hooks";
import type { StaffPickup } from "../schemas";

type StaffCounterReadyDialogProps = {
  /** The pickup whose counter part is about to be marked ready; null keeps the dialog closed. */
  pickup: StaffPickup | null;
  onClose: () => void;
};

/**
 * Confirms the counter's part of an order is ready, saying what that commits
 * the bakery to: an order still to start locks for the customer, and once
 * every station's part is ready the customer is told to come.
 */
export function StaffCounterReadyDialog({ pickup, onClose }: StaffCounterReadyDialogProps) {
  const markReady = usePatchCounterTicketStatus();
  const ticket = pickup?.counter_ticket ?? null;
  const locks =
    pickup?.status === "confirmed" ? " The customer can no longer cancel the order or move its pickup." : "";

  return (
    <ConfirmDialog
      confirmLabel="Mark ready"
      confirmVariant="warning"
      description={
        pickup && ticket
          ? `${pickup.code}: ${formatTicketItems(ticket.items)}.${locks} Once every station's part is ready, the customer is told the order can be collected.`
          : ""
      }
      isPending={markReady.isPending}
      open={ticket !== null}
      title="Counter part ready?"
      onCancel={onClose}
      onConfirm={() => {
        if (ticket) {
          markReady.mutate({ ticketId: ticket.id, status: "ready" }, { onSettled: onClose });
        }
      }}
    />
  );
}
