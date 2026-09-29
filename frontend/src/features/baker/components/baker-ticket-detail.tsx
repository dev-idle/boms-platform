"use client";

import { z } from "zod";
import { Button } from "@/components/ui/button";
import { InlineLoadingState } from "@/components/ui/loading-state";
import {
  formatOrderStatusLabel,
  StatusPill,
  ticketStatusToPillVariant,
} from "@/components/ui/status-pill";
import { isApiError } from "@/lib/errors";
import { nextTicketAction, STATION_LABEL } from "@/lib/schemas/ticket";
import { formatPickupDateTime } from "@/lib/validation/pickup";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";

import { useKitchenTicket, usePatchKitchenTicketStatus } from "../hooks";

type BakerTicketDetailProps = {
  ticketId: string;
};

/** A kitchen ticket to start and finish, beside where the rest of its order stands. */
export function BakerTicketDetail({ ticketId }: BakerTicketDetailProps) {
  const isValidId = z.uuid().safeParse(ticketId).success;
  const ticketQuery = useKitchenTicket(ticketId);
  const patchStatus = usePatchKitchenTicketStatus(ticketId);

  if (!isValidId) {
    return <p className="text-sm text-muted">Invalid ticket link.</p>;
  }

  if (ticketQuery.isPending) {
    return <InlineLoadingState />;
  }

  if (ticketQuery.isError) {
    const message =
      isApiError(ticketQuery.error) && ticketQuery.error.status === 404
        ? "Ticket not found."
        : "Failed to load the ticket.";
    return <p className="text-sm text-error">{message}</p>;
  }

  const ticket = ticketQuery.data;
  if (!ticket) {
    return <p className="text-sm text-muted">Ticket not found.</p>;
  }

  const action = nextTicketAction(ticket.status);

  return (
    <div className="dashboard-profile-section-stack">
      <DashboardProfileSection id="baker-ticket" title="Kitchen ticket">
        <div className="dashboard-order-summary">
          <div className="flex flex-wrap items-center gap-2">
            {ticket.pickup_at ? (
              <p className="text-sm text-muted">
                Pickup {formatPickupDateTime(ticket.pickup_at)}
              </p>
            ) : null}
            <StatusPill
              label={formatOrderStatusLabel(ticket.status)}
              variant={ticketStatusToPillVariant(ticket.status)}
            />
          </div>
          <p className="text-order-code">{ticket.order_code}</p>
          {ticket.customer.display_name ? (
            <p className="text-sm text-muted">
              Customer: {ticket.customer.display_name}
            </p>
          ) : null}

          <ul className="dashboard-order-line-items">
            {/* A product can arrive twice — on its own and inside a combo — so the
                line position is the key. */}
            {ticket.items.map((item, index) => (
              <li key={index} className="dashboard-order-line-item">
                <p className="font-medium text-ink">
                  {item.quantity}× {item.name}
                </p>
              </li>
            ))}
          </ul>

          <div className="dashboard-profile-form-actions">
            <Button
              aria-busy={patchStatus.isPending}
              disabled={!action || patchStatus.isPending}
              title={action ? undefined : "The ticket is ready for pickup"}
              type="button"
              onClick={() => {
                if (action) {
                  patchStatus.mutate({ status: action.status });
                }
              }}
            >
              {patchStatus.isPending ? "Updating…" : (action?.label ?? "Mark ready")}
            </Button>
          </div>
        </div>
      </DashboardProfileSection>

      <DashboardProfileSection
        description="The order is ready for pickup when every station's ticket is."
        id="baker-order-tickets"
        title="Whole order"
        variant="plain"
      >
        <ul className="dashboard-order-line-items">
          {ticket.order_tickets.map((orderTicket) => (
            <li key={orderTicket.station} className="dashboard-order-line-item">
              <p className="font-medium text-ink">{STATION_LABEL[orderTicket.station]}</p>
              <StatusPill
                label={formatOrderStatusLabel(orderTicket.status)}
                variant={ticketStatusToPillVariant(orderTicket.status)}
              />
            </li>
          ))}
        </ul>
      </DashboardProfileSection>
    </div>
  );
}
