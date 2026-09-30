import { describe, expect, it } from "vitest";

import {
  activeOrderProgressIndex,
  isOrderDropped,
  isWithTheBakery,
  statusReachedAt,
} from "./order-progress";

describe("order progress", () => {
  it("maps lifecycle statuses to step index", () => {
    expect(activeOrderProgressIndex("awaiting_payment")).toBe(0);
    expect(activeOrderProgressIndex("pending")).toBe(0);
    expect(activeOrderProgressIndex("confirmed")).toBe(1);
    expect(activeOrderProgressIndex("ready")).toBe(3);
    expect(activeOrderProgressIndex("fulfilled")).toBe(4);
  });

  it("marks orders that will not be made separately", () => {
    expect(activeOrderProgressIndex("cancelled")).toBe(-1);
    expect(activeOrderProgressIndex("expired")).toBe(-1);
    expect(isOrderDropped("cancelled")).toBe(true);
    expect(isOrderDropped("expired")).toBe(true);
    expect(isOrderDropped("awaiting_payment")).toBe(false);
  });

  it("shows the stations only once the bakery has the order", () => {
    expect(isWithTheBakery("awaiting_payment")).toBe(false);
    expect(isWithTheBakery("cancelled")).toBe(false);
    expect(isWithTheBakery("expired")).toBe(false);
    expect(isWithTheBakery("confirmed")).toBe(true);
    expect(isWithTheBakery("ready")).toBe(true);
  });

  it("reads when each step was reached from the timeline", () => {
    const timeline = [
      { status: "pending", at: "2026-09-28T03:00:00Z" },
      { status: "confirmed", at: "2026-09-28T03:05:00Z" },
    ] as const;

    expect(statusReachedAt(timeline, "pending")).toBe("2026-09-28T03:00:00Z");
    expect(statusReachedAt(timeline, "confirmed")).toBe("2026-09-28T03:05:00Z");
    expect(statusReachedAt(timeline, "ready")).toBeUndefined();
  });
});
