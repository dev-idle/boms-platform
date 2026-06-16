export const COMBO_BUNDLE_MIN_MESSAGE =
  "Add at least two products, or one product with quantity of at least 2";

/** Combo must include ≥2 products or one product with quantity ≥2. */
export function isValidComboBundle(
  items: ReadonlyArray<{ quantity: number }>,
): boolean {
  if (items.length >= 2) {
    return true;
  }
  if (items.length === 1 && items[0].quantity >= 2) {
    return true;
  }
  return false;
}
