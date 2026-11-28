"use client";

import Link from "next/link";

import { ROUTE } from "@/constants/routes";
import { CATALOG_COMBOS_HEADING_ID, CUSTOM_CAKE_BROWSE_HREF, useCombosOnSale } from "@/features/catalog";

/** The footer's shop links: Combos only while any are on sale, as in the header. */
export function StorefrontFooterShopLinks() {
  const combosOnSale = useCombosOnSale();

  return (
    <ul className="storefront-footer__links">
      <li>
        <Link href={ROUTE.products}>All products</Link>
      </li>
      {combosOnSale ? (
        <li>
          <Link href={`${ROUTE.products}#${CATALOG_COMBOS_HEADING_ID}`}>Combos</Link>
        </li>
      ) : null}
      <li>
        <Link href={CUSTOM_CAKE_BROWSE_HREF}>Custom cakes</Link>
      </li>
    </ul>
  );
}
