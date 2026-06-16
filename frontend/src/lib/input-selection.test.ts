import { describe, expect, it } from "vitest";

import {
  clampSelection,
  isUntrustedPasswordSelectionRead,
  rememberTrustedSelectionRead,
} from "./input-selection";

describe("clampSelection", () => {
  it("clamps indices to the value length", () => {
    expect(clampSelection({ start: -2, end: 9 }, 5)).toEqual({
      start: 0,
      end: 5,
    });
  });
});

describe("isUntrustedPasswordSelectionRead", () => {
  it("flags 0/0 on a non-empty password field", () => {
    expect(
      isUntrustedPasswordSelectionRead("password", 6, { start: 0, end: 0 }),
    ).toBe(true);
  });

  it("accepts a real caret position", () => {
    expect(
      isUntrustedPasswordSelectionRead("password", 6, { start: 3, end: 3 }),
    ).toBe(false);
  });
});

describe("rememberTrustedSelectionRead", () => {
  it("keeps the previous caret when password selection is untrusted", () => {
    const previous = { start: 3, end: 3 };
    expect(
      rememberTrustedSelectionRead("password", 3, { start: 0, end: 0 }, previous),
    ).toEqual(previous);
  });
});
