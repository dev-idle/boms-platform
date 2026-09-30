import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

import { TERMS_VERSION } from "./policies";

// contracts/terms-version.json, which the backend's TermsVersion is tested
// against too: the pages and the API must name the same policies.
describe("TERMS_VERSION", () => {
  it("is the version the API accepts", () => {
    const contract = JSON.parse(
      readFileSync(new URL("../../../contracts/terms-version.json", import.meta.url), "utf8"),
    ) as { version: string };

    expect(TERMS_VERSION).toBe(contract.version);
  });
});
