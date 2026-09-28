import { describe, expect, it } from "vitest";

import {
  activeOrderProgressIndex,
  isOrderCancelled,
  statusReachedAt,
} from "./order-progress";

describe("order progress", () => {
  it("maps lifecycle statuses to step index", () => {
    expect(activeOrderProgressIndex("pending")).toBe(0);
    expect(activeOrderProgressIndex("ready")).toBe(3);
    expect(activeOrderProgressIndex("fulfilled")).toBe(4);
  });

  it("marks cancelled orders separately", () => {
    expect(activeOrderProgressIndex("cancelled")).toBe(-1);
    expect(isOrderCancelled("cancelled")).toBe(true);
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
