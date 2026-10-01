import { describe, expect, it } from "vitest";

import {
  formatTicketItems,
  nextTicketAction,
  stationTicketSchema,
  ticketChangeMessage,
  ticketItemCount,
} from "./ticket";

const ticket = {
  id: "00000000-0000-4000-8000-000000000001",
  order_id: "00000000-0000-4000-8000-000000000002",
  order_code: "CH-260928-001",
  order_status: "confirmed",
  station: "kitchen",
  status: "queued",
  pickup_at: "2026-09-29T09:00:00+07:00",
  items: [{ name: "Matcha cake", quantity: 2, customization: null }],
  created_at: "2026-09-28T09:00:00+07:00",
};

// A station API sends a name for the order and no way to reach the customer.
describe("station ticket customer", () => {
  it("reads a ticket that carries only the customer's display name", () => {
    expect(stationTicketSchema.safeParse({ ...ticket, customer: { display_name: "Mai" } }).success).toBe(true);
  });

  it("reads a ticket whose customer gave no name", () => {
    expect(stationTicketSchema.safeParse({ ...ticket, customer: {} }).success).toBe(true);
  });

  it("drops contact details if a response ever carried them", () => {
    const result = stationTicketSchema.parse({
      ...ticket,
      customer: { display_name: "Mai", email: "mai@example.com" },
    });
    expect(result.customer).toEqual({ display_name: "Mai" });
  });

  it("rejects a cancelled ticket, which has left the queue", () => {
    expect(stationTicketSchema.safeParse({ ...ticket, customer: {}, status: "cancelled" }).success).toBe(false);
  });
});

describe("nextTicketAction", () => {
  it("starts a queued ticket and finishes a started one", () => {
    expect(nextTicketAction("queued")).toEqual({ label: "Start", status: "in_progress" });
    expect(nextTicketAction("in_progress")).toEqual({ label: "Mark ready", status: "ready" });
  });

  it("offers nothing once a ticket is ready or cancelled", () => {
    expect(nextTicketAction("ready")).toBeNull();
    expect(nextTicketAction("cancelled")).toBeNull();
  });
});

describe("ticket items", () => {
  const items = [
    { name: "Matcha cake", quantity: 2, customization: null },
    { name: "Croissant", quantity: 1, customization: null },
  ];

  it("reads as one line", () => {
    expect(formatTicketItems(items)).toBe("2× Matcha cake · 1× Croissant");
  });

  it("counts every product", () => {
    expect(ticketItemCount(items)).toBe(3);
  });
});

describe("ticketChangeMessage", () => {
  const change = {
    id: "00000000-0000-4000-8000-000000000001",
    order_id: "00000000-0000-4000-8000-000000000002",
    station: "kitchen" as const,
  };

  it("says when the last ticket finished the order", () => {
    expect(ticketChangeMessage({ ...change, status: "ready", order_status: "ready" })).toBe(
      "Ticket ready — the whole order is ready for pickup",
    );
    expect(ticketChangeMessage({ ...change, status: "ready", order_status: "in_production" })).toBe("Ticket ready");
    expect(ticketChangeMessage({ ...change, status: "in_progress", order_status: "in_production" })).toBe(
      "Ticket started",
    );
  });
});
