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

/** Scroll offset for storefront in-page anchors (header + breathing room). */

export function getStorefrontAnchorScrollOffset(): number {
  const root = document.documentElement;
  const styles = getComputedStyle(root);
  const headerHeight =
    Number.parseFloat(styles.getPropertyValue("--storefront-header-height")) ||
    72;
  const anchorGap =
    Number.parseFloat(styles.getPropertyValue("--storefront-anchor-gap")) || 56;

  let offset = headerHeight + anchorGap;

  if (window.matchMedia("(max-width: 63.99rem)").matches) {
    const stickyMenu = document.querySelector(".catalog-menu");
    if (stickyMenu instanceof HTMLElement) {
      offset += stickyMenu.offsetHeight;
    }
  }

  return offset;
}

/** Tighter offset so combo cards stay in view after menu navigation. */
export function getStorefrontCombosScrollOffset(): number {
  const root = document.documentElement;
  const styles = getComputedStyle(root);
  const headerHeight =
    Number.parseFloat(styles.getPropertyValue("--storefront-header-height")) ||
    72;
  const anchorGap =
    Number.parseFloat(styles.getPropertyValue("--storefront-combos-anchor-gap")) ||
    20;

  let offset = headerHeight + anchorGap;

  if (window.matchMedia("(max-width: 63.99rem)").matches) {
    const stickyMenu = document.querySelector(".catalog-menu");
    if (stickyMenu instanceof HTMLElement) {
      offset += stickyMenu.offsetHeight;
    }
  }

  return offset;
}

export function scrollToStorefrontAnchor(
  elementId: string,
  offset = getStorefrontAnchorScrollOffset(),
): void {
  const target = document.getElementById(elementId);
  if (!target) {
    return;
  }

  const top =
    target.getBoundingClientRect().top + window.scrollY - offset;

  window.scrollTo({
    top: Math.max(0, top),
    behavior: "smooth",
  });
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
