import { describe, expect, it } from "vitest";

import { parseOperationalEmployeeCode } from "./employee-code";

describe("parseOperationalEmployeeCode", () => {
  it("parses standard EMP codes", () => {
    expect(parseOperationalEmployeeCode("EMP-00001")).toEqual({
      code: "EMP-00001",
      sequence: 1,
    });
    expect(parseOperationalEmployeeCode("EMP-00042")).toEqual({
      code: "EMP-00042",
      sequence: 42,
    });
  });

  it("rejects non-standard codes", () => {
    expect(parseOperationalEmployeeCode("STAFF-1")).toBeNull();
    expect(parseOperationalEmployeeCode("")).toBeNull();
  });
});
