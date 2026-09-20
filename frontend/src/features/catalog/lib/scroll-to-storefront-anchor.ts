/** In-page anchor for the combo section title on `/products`. */
export const CATALOG_COMBOS_HEADING_ID = "catalog-combos-heading";

/** Legacy scroll hash — resolves to the header anchor. */
export const CATALOG_COMBOS_SCROLL_ANCHOR_ID = "catalog-combos-scroll";

/** Legacy section id — old bookmarks used `#combos`. */
export const CATALOG_COMBOS_SECTION_ID = "combos";

export function resolveStorefrontCombosHeadingHash(hash?: string): string | null {
  const normalized =
    hash ??
    (typeof window !== "undefined" ? window.location.hash.slice(1) : "");
  if (
    normalized === CATALOG_COMBOS_HEADING_ID ||
    normalized === CATALOG_COMBOS_SECTION_ID ||
    normalized === CATALOG_COMBOS_SCROLL_ANCHOR_ID
  ) {
    return CATALOG_COMBOS_HEADING_ID;
  }
  return null;
}

export function scrollToStorefrontCombosSection(): void {
  const target = document.getElementById(CATALOG_COMBOS_HEADING_ID);
  if (!target) {
    return;
  }

  const behavior = window.matchMedia("(prefers-reduced-motion: reduce)").matches
    ? "auto"
    : "smooth";

  target.scrollIntoView({ behavior, block: "start" });
}
