"use client";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { DashboardTableActionButton } from "@/components/ui/dashboard-table-action-button";
import {
  formatOrderStatusLabel,
  StatusPill,
  ticketStatusToPillVariant,
} from "@/components/ui/status-pill";
import type { OrderStatus } from "@/lib/schemas/order";
import { formatTicketItems, STATION_LABEL } from "@/lib/schemas/ticket";

import { useMoveTicket } from "../hooks";
import { otherStation, ticketMoveBlockedReason } from "../lib/ticket-move";
import type { StaffOrderTicket } from "../schemas";

type StaffOrderTicketsProps = {
  orderId: string;
  orderStatus: OrderStatus;
  tickets: StaffOrderTicket[];
};

/** What each station makes for the order, and the move of a ticket nobody started. */
export function StaffOrderTickets({ orderId, orderStatus, tickets }: StaffOrderTicketsProps) {
  const move = useMoveTicket(orderId);

  // Orders handed over before tickets existed have none.
  if (tickets.length === 0) {
    return null;
  }

  return (
    <DashboardProfileSection
      description="A ticket nobody has started can move to the other station."
      id="staff-order-tickets"
      title="Preparation"
      variant="plain"
    >
      <ul className="dashboard-order-line-items">
        {tickets.map((ticket) => {
          const to = otherStation(ticket.station);
          const isMoving = move.isPending && move.variables?.ticketId === ticket.id;
          return (
            <li key={ticket.id} className="dashboard-order-line-item dashboard-order-ticket">
              <div className="dashboard-order-ticket__items">
                <p className="font-medium text-ink">{STATION_LABEL[ticket.station]}</p>
                <p className="text-muted">{formatTicketItems(ticket.items)}</p>
              </div>
              <div className="dashboard-order-ticket__state">
                <StatusPill
                  label={formatOrderStatusLabel(ticket.status)}
                  variant={ticketStatusToPillVariant(ticket.status)}
                />
                <DashboardTableActionButton
                  blockedReason={
                    isMoving
                      ? "Moving the ticket"
                      : ticketMoveBlockedReason(ticket, tickets, orderStatus)
                  }
                  label={`Move to ${STATION_LABEL[to].toLowerCase()} — the ${STATION_LABEL[ticket.station].toLowerCase()} ticket`}
                  onClick={() => move.mutate({ ticketId: ticket.id, station: to })}
                  text={`Move to ${STATION_LABEL[to].toLowerCase()}`}
                  tone="accent"
                />
              </div>
            </li>
          );
        })}
      </ul>
    </DashboardProfileSection>
  );
}
