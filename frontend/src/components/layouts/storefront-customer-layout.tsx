"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { StorefrontBrowseLink } from "@/components/layouts/storefront-browse-link";
import {
  isStorefrontCustomerNavActive,
  STOREFRONT_CUSTOMER_NAV,
} from "@/constants/storefront-customer-nav";
import { STOREFRONT_NAV_COPY } from "@/constants/storefront-nav-copy";
import { ROUTE } from "@/constants/routes";
import { cn } from "@/lib/utils";

type StorefrontCustomerLayoutProps = {
  children: ReactNode;
};

function customerBackNavForPath(pathname: string): {
  href: string;
  label: string;
} {
  if (pathname.startsWith("/orders/") && pathname !== ROUTE.orders) {
    return { href: ROUTE.orders, label: STOREFRONT_NAV_COPY.returnToOrders };
  }

  return { href: ROUTE.products, label: STOREFRONT_NAV_COPY.returnToShop };
}

/** Customer area shell — contextual nav for cart, orders, and account. */
export function StorefrontCustomerLayout({
  children,
}: StorefrontCustomerLayoutProps) {
  const pathname = usePathname();
  const backNav = customerBackNavForPath(pathname);

  return (
    <div className="storefront-customer-layout">
      <nav aria-label="Your account" className="storefront-customer-nav">
        <p className="storefront-customer-nav__eyebrow">Your shop</p>
        <ul className="storefront-customer-nav__list">
          {STOREFRONT_CUSTOMER_NAV.map((item) => {
            const active = isStorefrontCustomerNavActive(pathname, item.href);
            const Icon = item.Icon;

            return (
              <li key={item.href}>
                <Link
                  aria-current={active ? "page" : undefined}
                  className={cn(
                    "storefront-customer-nav__link",
                    active && "storefront-customer-nav__link--active",
                  )}
                  href={item.href}
                >
                  <Icon className="storefront-customer-nav__icon" />
                  <span className="storefront-customer-nav__label">{item.label}</span>
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>

      <div className="storefront-customer-main">
        <StorefrontBrowseLink href={backNav.href} withNavShell>
          {backNav.label}
        </StorefrontBrowseLink>
        {children}
      </div>
    </div>
  );
}
