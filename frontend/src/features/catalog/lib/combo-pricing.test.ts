import { describe, expect, it } from "vitest";

import type { CatalogCombo } from "@/lib/schemas/catalog";

import {
  comboRetailTotalCents,
  comboSavingsCents,
} from "./combo-pricing";

const sampleCombo: CatalogCombo = {
  id: "00000000-0000-4000-8000-000000000001",
  name: "Morning set",
  slug: "morning-set",
  price_cents: 1200,
  starts_at: "2026-01-01T00:00:00+00:00",
  ends_at: "2026-12-31T23:59:59+00:00",
  items: [
    {
      product_id: "00000000-0000-4000-8000-000000000002",
      product_name: "Croissant",
      product_slug: "croissant",
      quantity: 2,
      price_cents: 500,
    },
    {
      product_id: "00000000-0000-4000-8000-000000000003",
      product_name: "Coffee",
      product_slug: "coffee",
      quantity: 1,
      price_cents: 400,
    },
  ],
};

describe("combo-pricing", () => {
  it("sums retail total from line items", () => {
    expect(comboRetailTotalCents(sampleCombo.items)).toBe(1400);
  });

  it("computes non-negative savings", () => {
    expect(comboSavingsCents(sampleCombo)).toBe(200);
  });

  it("never returns negative savings when bundle costs more", () => {
    expect(
      comboSavingsCents({
        ...sampleCombo,
        price_cents: 2000,
      }),
    ).toBe(0);
  });
});
