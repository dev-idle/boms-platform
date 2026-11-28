import Link from "next/link";

import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

export type StorefrontBackTarget = {
  href: string;
  label: string;
};

/** The whole shop, the step above a category. */
export const SHOP_BACK: StorefrontBackTarget = { href: ROUTE.products, label: PAGE_TITLES.shop };

/**
 * The one step up from a page: a product's category, a category's shop, an
 * order's list. The shop is small, and its header already reaches the rest, so
 * a page names only the place it came from.
 */
export function StorefrontBackLink({ href, label }: StorefrontBackTarget) {
  return (
    <Link className="storefront-back-link" href={href}>
      <span aria-hidden="true" className="storefront-back-link__arrow">
        ←
      </span>
      {label}
    </Link>
  );
}
