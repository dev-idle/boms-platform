import { describe, expect, it, vi } from "vitest";

import { takeLinkToken } from "./link-token";

function page(hash: string) {
  const history = { replaceState: vi.fn() };
  return { location: { hash, pathname: "/reset-password", search: "" }, history };
}

describe("takeLinkToken", () => {
  it("reads the token and clears the fragment", () => {
    const { location, history } = page("#token=Abc_def-1234567890XYZ");
    expect(takeLinkToken(location, history)).toBe("Abc_def-1234567890XYZ");
    expect(history.replaceState).toHaveBeenCalledWith(null, "", "/reset-password");
  });

  it("refuses a missing or malformed token", () => {
    expect(takeLinkToken(page("").location, page("").history)).toBeNull();
    expect(takeLinkToken(page("#token=short").location, page("#token=short").history)).toBeNull();
    const odd = page("#token=%3Cscript%3E0123456789abcdef");
    expect(takeLinkToken(odd.location, odd.history)).toBeNull();
    expect(odd.history.replaceState).toHaveBeenCalled();
  });
});
