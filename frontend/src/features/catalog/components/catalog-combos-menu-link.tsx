"use client";

import {
  CATALOG_COMBOS_HEADING_ID,
  scrollToStorefrontCombosSection,
} from "../lib/scroll-to-storefront-anchor";

export function CatalogCombosMenuLink() {
  return (
    <a
      className="catalog-chip"
      href={`#${CATALOG_COMBOS_HEADING_ID}`}
      onClick={(event) => {
        event.preventDefault();
        scrollToStorefrontCombosSection();
        window.history.replaceState(null, "", `#${CATALOG_COMBOS_HEADING_ID}`);
      }}
    >
      Combos
    </a>
  );
}
