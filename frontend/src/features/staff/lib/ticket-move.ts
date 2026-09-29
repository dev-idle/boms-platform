import type { OrderStatus } from "@/lib/schemas/order";
import { STATION_LABEL, type Station } from "@/lib/schemas/ticket";

import type { StaffOrderTicket } from "../schemas";

/** The station a ticket would move to. */
export function otherStation(station: Station): Station {
  return station === "kitchen" ? "counter" : "kitchen";
}

/**
 * Why the counter cannot move this ticket, or undefined when it can. These are
 * the API's rules: the order is still open, nobody has started the ticket, and
 * the other station has no ticket of this order yet.
 */
export function ticketMoveBlockedReason(
  ticket: StaffOrderTicket,
  orderTickets: ReadonlyArray<StaffOrderTicket>,
  orderStatus: OrderStatus,
): string | undefined {
  if (orderStatus === "cancelled" || orderStatus === "fulfilled") {
    return "The order is closed";
  }
  if (ticket.status !== "queued") {
    return "Only a ticket nobody has started can move";
  }
  const to = otherStation(ticket.station);
  if (orderTickets.some((other) => other.station === to)) {
    return `The ${STATION_LABEL[to].toLowerCase()} already has a ticket for this order`;
  }
  return undefined;
}
