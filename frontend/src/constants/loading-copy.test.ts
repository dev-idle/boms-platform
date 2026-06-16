import { describe, expect, it } from "vitest";

import { LOADING_MESSAGE } from "./loading-copy";

describe("loading-copy", () => {
  it("uses one user-facing message", () => {
    expect(LOADING_MESSAGE).toBe("Loading…");
  });
});
