import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

import { BRAND } from "./brand";

// contracts/brand.json, which the backend's email templates are tested against
// too: a customer reads the same address and phone on the site and in an email.
describe("BRAND", () => {
  it("shows the contact details the emails show", () => {
    const contract = JSON.parse(
      readFileSync(new URL("../../../contracts/brand.json", import.meta.url), "utf8"),
    ) as { name: string; address_line: string; contact_email: string; contact_phone: string };

    expect({
      name: BRAND.name,
      address_line: BRAND.addressLine,
      contact_email: BRAND.contactEmail,
      contact_phone: BRAND.contactPhone,
    }).toEqual({
      name: contract.name,
      address_line: contract.address_line,
      contact_email: contract.contact_email,
      contact_phone: contract.contact_phone,
    });
  });
});
