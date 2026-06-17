import type { CatalogCombo } from "@/lib/schemas/catalog";

export function comboRetailTotalCents(
  items: CatalogCombo["items"],
): number {
  return items.reduce(
    (sum, item) => sum + item.price_cents * item.quantity,
    0,
  );
}

export function comboSavingsCents(combo: CatalogCombo): number {
  return Math.max(0, comboRetailTotalCents(combo.items) - combo.price_cents);
}

export function formatComboEndsAt(iso: string): string {
  return new Intl.DateTimeFormat("en-US", {
    day: "numeric",
    month: "short",
  }).format(new Date(iso));
}
