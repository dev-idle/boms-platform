import type { ReactNode } from "react";
import Link from "next/link";

import { ROUTE } from "@/constants/routes";
import { STOREFRONT_NAV_COPY } from "@/constants/storefront-nav-copy";
import { cn } from "@/lib/utils";

type StorefrontBrowseLinkProps = {
  children?: ReactNode;
  className?: string;
  href?: string;
  /** Wrap in nav shell for customer / catalog back rows. */
  withNavShell?: boolean;
};

/** Storefront exit link — same chrome as login (`auth-page-exit-link`). */
export function StorefrontBrowseLink({
  children = STOREFRONT_NAV_COPY.returnToShop,
  className,
  href = ROUTE.products,
  withNavShell = false,
}: StorefrontBrowseLinkProps) {
  const link = (
    <Link className={cn("auth-page-exit-link", className)} href={href}>
      <span aria-hidden="true" className="auth-page-exit-icon">
        ←
      </span>
      <span className="auth-page-exit-label">{children}</span>
    </Link>
  );

  if (!withNavShell) {
    return link;
  }

  return (
    <nav aria-label="Back" className="storefront-back-nav">
      {link}
    </nav>
  );
}
