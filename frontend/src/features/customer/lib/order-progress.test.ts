import { describe, expect, it } from "vitest";

import {
  activeOrderProgressIndex,
  isUncollected,
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

  it("marks orders that end without being collected separately", () => {
    for (const status of ["cancelled", "expired", "no_show"] as const) {
      expect(activeOrderProgressIndex(status)).toBe(-1);
      expect(isUncollected(status)).toBe(true);
    }
    expect(isUncollected("awaiting_payment")).toBe(false);
    expect(isUncollected("fulfilled")).toBe(false);
  });

  it("shows the stations only once the bakery has the order", () => {
    expect(isWithTheBakery("awaiting_payment")).toBe(false);
    expect(isWithTheBakery("cancelled")).toBe(false);
    expect(isWithTheBakery("expired")).toBe(false);
    expect(isWithTheBakery("no_show")).toBe(false);
    expect(isWithTheBakery("confirmed")).toBe(true);
    expect(isWithTheBakery("ready")).toBe(true);
  });

  it("reads when each step was reached from the timeline", () => {
    const timeline = [
      { status: "pending", reason: null, at: "2026-09-28T03:00:00Z" },
      { status: "confirmed", reason: null, at: "2026-09-28T03:05:00Z" },
    ] as const;

    expect(statusReachedAt(timeline, "pending")).toBe("2026-09-28T03:00:00Z");
    expect(statusReachedAt(timeline, "confirmed")).toBe("2026-09-28T03:05:00Z");
    expect(statusReachedAt(timeline, "ready")).toBeUndefined();
  });
});
