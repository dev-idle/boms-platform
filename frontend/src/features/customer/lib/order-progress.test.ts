import { describe, expect, it } from "vitest";

import {
  activeOrderProgressIndex,
  isOrderCancelled,
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
});
