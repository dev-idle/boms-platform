"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState, type ReactNode } from "react";

import { ROUTE } from "@/constants/routes";
import { SignOutDialog } from "@/features/auth";
import { useOrdersGlance } from "@/features/customer";
import { isNavItemActive } from "@/lib/routing/nav-active";
import { plural } from "@/lib/utils";

import { StorefrontPageHeader } from "./storefront-page-header";

const ACCOUNT_LINKS = [
  { href: ROUTE.orders, label: "Orders" },
  { href: ROUTE.customer.saved, label: "Saved" },
  { href: ROUTE.customer.account.profile, label: "Account" },
] as const;

type StorefrontAccountShellProps = {
  children: ReactNode;
  title: string;
};

/** One of the customer's own pages: its title and content beside the menu of all of them. */
export function StorefrontAccountShell({ children, title }: StorefrontAccountShellProps) {
  return (
    <div className="storefront-subnav-shell">
      <StorefrontPageHeader title={title} />
      <StorefrontAccountNav />
      <div className="storefront-subnav-shell__main">{children}</div>
    </div>
  );
}

function StorefrontAccountNav() {
  const pathname = usePathname();
  const unread = useOrdersGlance(true).data?.unread ?? 0;
  const [signingOut, setSigningOut] = useState(false);

  return (
    <nav aria-label="Your account" className="storefront-subnav">
      <ul className="storefront-subnav__list">
        {ACCOUNT_LINKS.map((link) => (
          <li key={link.href}>
            <Link
              aria-current={isNavItemActive(pathname, link.href, "exact") ? "page" : undefined}
              className="storefront-subnav__link"
              href={link.href}
            >
              {link.label}{" "}
              {link.href === ROUTE.orders && unread > 0 ? (
                <span className="storefront-subnav__meta">{plural(unread, "new message")}</span>
              ) : null}
            </Link>
          </li>
        ))}
      </ul>
      <button className="storefront-subnav__signout" type="button" onClick={() => setSigningOut(true)}>
        Sign out
      </button>
      <SignOutDialog open={signingOut} onClose={() => setSigningOut(false)} />
    </nav>
  );
}
