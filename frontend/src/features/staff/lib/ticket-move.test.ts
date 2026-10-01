import { describe, expect, it } from "vitest";

import type { StaffOrderTicket } from "../schemas";
import { otherStation, ticketMoveBlockedReason } from "./ticket-move";

const ticket = (station: StaffOrderTicket["station"], status: StaffOrderTicket["status"]): StaffOrderTicket => ({
  id: `00000000-0000-4000-8000-00000000000${station === "kitchen" ? 1 : 2}`,
  station,
  status,
  items: [{ name: "Croissant", quantity: 1, customization: null }],
});

describe("otherStation", () => {
  it("swaps the kitchen and the counter", () => {
    expect(otherStation("kitchen")).toBe("counter");
    expect(otherStation("counter")).toBe("kitchen");
  });
});

describe("ticketMoveBlockedReason", () => {
  it("lets a queued ticket move to a free station of an open order", () => {
    const counter = ticket("counter", "queued");
    expect(ticketMoveBlockedReason(counter, [counter], "confirmed")).toBeUndefined();
  });

  it("keeps a started ticket where it is", () => {
    const counter = ticket("counter", "in_progress");
    expect(ticketMoveBlockedReason(counter, [counter], "in_production")).toBe(
      "Only a ticket nobody has started can move",
    );
  });

  it("keeps one ticket per station", () => {
    const counter = ticket("counter", "queued");
    expect(ticketMoveBlockedReason(counter, [counter, ticket("kitchen", "queued")], "confirmed")).toBe(
      "The kitchen already has a ticket for this order",
    );
  });

  it("moves nothing on a closed order", () => {
    const counter = ticket("counter", "queued");
    for (const status of ["cancelled", "fulfilled"] as const) {
      expect(ticketMoveBlockedReason(counter, [counter], status)).toBe("The order is closed");
    }
  });
});
