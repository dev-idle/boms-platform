"use client";

import { cn } from "@/lib/utils";

import {
  CATALOG_COMBOS_HEADING_ID,
  scrollToStorefrontCombosSection,
} from "../lib/scroll-to-storefront-anchor";

type CatalogCombosMenuLinkProps = {
  className?: string;
};

export function CatalogCombosMenuLink({ className }: CatalogCombosMenuLinkProps) {
  return (
    <a
      className={cn("catalog-menu__link catalog-menu__link--anchor", className)}
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
