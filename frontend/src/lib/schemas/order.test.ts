import { describe, expect, it } from "vitest";

import { orderCodeSchema } from "./order";

describe("orderCodeSchema", () => {
  it("reads the codes checkout hands out", () => {
    for (const code of ["CH-260928-001", "CH-260928-042", "CH-260928-1000"]) {
      expect(orderCodeSchema.safeParse(code).success, code).toBe(true);
    }
  });

  it("rejects anything else", () => {
    for (const code of ["", "CH-260928-01", "ch-260928-001", "CH-2609-001", "00000000"]) {
      expect(orderCodeSchema.safeParse(code).success, code).toBe(false);
    }
  });
});
