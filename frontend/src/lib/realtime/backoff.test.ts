import { describe, expect, it } from "vitest";

import { reconnectDelay } from "./backoff";

describe("reconnectDelay", () => {
  it("doubles the wait per attempt within a jittered half", () => {
    expect(reconnectDelay(0, () => 0)).toBe(500);
    expect(reconnectDelay(0, () => 1)).toBe(1_000);
    expect(reconnectDelay(3, () => 0)).toBe(4_000);
    expect(reconnectDelay(3, () => 1)).toBe(8_000);
  });

  it("never waits longer than 30 seconds", () => {
    expect(reconnectDelay(20, () => 1)).toBe(30_000);
    expect(reconnectDelay(20, () => 0)).toBe(15_000);
  });
});
