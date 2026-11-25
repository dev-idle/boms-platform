import { describe, expect, it } from "vitest";

import { kitchenTicketSchema, nextTicketAction, stationTicketSchema, ticketItemCount } from "./index";

const ticket = {
  id: "00000000-0000-4000-8000-000000000001",
  order_id: "00000000-0000-4000-8000-000000000002",
  order_code: "CH-260928-001",
  order_status: "in_production",
  station: "kitchen",
  status: "in_progress",
  customer: { display_name: "Mai" },
  items: [{ name: "Matcha cake", quantity: 1, customization: null }],
  created_at: "2026-09-28T09:00:00+07:00",
};

describe("kitchen ticket", () => {
  it("reads where the whole order stands", () => {
    const result = kitchenTicketSchema.parse({
      ...ticket,
      order_tickets: [
        { station: "kitchen", status: "in_progress" },
        { station: "counter", status: "ready" },
      ],
    });
    expect(result.order_tickets).toHaveLength(2);
  });

  it("rejects a station it does not know", () => {
    expect(
      kitchenTicketSchema.safeParse({ ...ticket, order_tickets: [{ station: "bar", status: "ready" }] }).success,
    ).toBe(false);
  });
});

const queued = {
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
    expect(stationTicketSchema.safeParse({ ...queued, customer: { display_name: "Mai" } }).success).toBe(true);
  });

  it("reads a ticket whose customer gave no name", () => {
    expect(stationTicketSchema.safeParse({ ...queued, customer: {} }).success).toBe(true);
  });

  it("drops contact details if a response ever carried them", () => {
    const result = stationTicketSchema.parse({
      ...queued,
      customer: { display_name: "Mai", email: "mai@example.com" },
    });
    expect(result.customer).toEqual({ display_name: "Mai" });
  });

  it("rejects a cancelled ticket, which has left the queue", () => {
    expect(stationTicketSchema.safeParse({ ...queued, customer: {}, status: "cancelled" }).success).toBe(false);
  });
});

describe("nextTicketAction", () => {
  it("starts a queued ticket and finishes a started one, saying what each commits the bakery to", () => {
    expect(nextTicketAction("queued")).toMatchObject({ label: "Start", status: "in_progress" });
    expect(nextTicketAction("queued")?.consequence).toContain("can no longer cancel");
    expect(nextTicketAction("in_progress")).toMatchObject({ label: "Mark ready", title: "Mark this ticket ready?", status: "ready" });
  });

  it("offers nothing once a ticket is ready or cancelled", () => {
    expect(nextTicketAction("ready")).toBeNull();
    expect(nextTicketAction("cancelled")).toBeNull();
  });
});


describe("ticketItemCount", () => {
  it("counts every product", () => {
    expect(
      ticketItemCount([
        { name: "Matcha cake", quantity: 2, customization: null },
        { name: "Croissant", quantity: 1, customization: null },
      ]),
    ).toBe(3);
  });
});
