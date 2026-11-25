import { describe, expect, it } from "vitest";

import { formatTicketItems, ticketChangeMessage } from "./ticket";

describe("ticket items", () => {
  const items = [
    { name: "Matcha cake", quantity: 2, customization: null },
    { name: "Croissant", quantity: 1, customization: null },
  ];

  it("reads as one line", () => {
    expect(formatTicketItems(items)).toBe("2× Matcha cake · 1× Croissant");
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
