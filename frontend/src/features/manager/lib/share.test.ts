import { describe, expect, it } from "vitest";

import { shareOf } from "./share";

describe("shareOf", () => {
  it("rounds down to a whole percentage", () => {
    expect(shareOf(3, 7)).toBe("42%");
    expect(shareOf(2, 3)).toBe("66%");
    expect(shareOf(599, 1000), "59.9% has not reached 60%").toBe("59%");
    expect(shareOf(7, 7)).toBe("100%");
    expect(shareOf(0, 7)).toBe("0%");
  });

  it("shows a dash when there is no customer yet", () => {
    expect(shareOf(0, 0)).toBe("—");
  });
});
