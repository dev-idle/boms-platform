"use client";

import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useRef, useState } from "react";

import { LogOutIcon } from "@/components/icons/dashboard-nav-icons";
import { ChevronDownIcon, UserIcon } from "@/components/icons/storefront-icons";
import { USER_ROLE } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";
import { SignOutDialog } from "@/features/auth";
import { useOrdersGlance } from "@/features/customer";
import { isNavItemActive } from "@/lib/routing/nav-active";
import { cn, plural } from "@/lib/utils";
import { formatPickupDateTime } from "@/lib/validation/pickup";
import { initialOf } from "@/lib/user-initials";
import { useAuthStore } from "@/stores/auth-store";

/** Where a signed-in customer's own pages live, which the trigger marks as current. */
const CUSTOMER_ROUTES = [ROUTE.orders, "/customer"] as const;

/** The first word of the name the customer gave; without one, the name part of their email. */
function accountName(displayName: string | undefined, email: string): string {
  return displayName?.split(/\s+/)[0] || email.replace(/@.*$/, "");
}

/**
 * Account at the end of the header — a guest goes to sign in; a signed-in
 * customer sees their monogram and name, so the header says who is signed in,
 * and opens a menu of their own pages with the next order to collect.
 */
export function StorefrontHeaderAccount() {
  const pathname = usePathname();
  const restoring = useAuthStore((state) => state.status === "idle");
  const user = useAuthStore((state) => (state.status === "authenticated" ? state.user : null));
  const customer = user?.role === USER_ROLE.customer ? user : null;
  const glance = useOrdersGlance(customer !== null).data;
  const [signingOut, setSigningOut] = useState(false);
  // Asked for from the menu, and opened once the menu has closed and given
  // focus back to its trigger — where the dialog returns it when it closes.
  const signOutAsked = useRef(false);

  // While a session is restored the header cannot say who is signed in, so it
  // shows neither Sign in nor a name until it knows.
  if (restoring) {
    return (
      <span aria-hidden="true" className="storefront-header-account storefront-header-account--pending">
        <UserIcon />
      </span>
    );
  }

  if (!customer) {
    return (
      <Link className="storefront-header-account" href={ROUTE.login}>
        <UserIcon />
        <span className="sr-only lg:not-sr-only">Sign in</span>
      </Link>
    );
  }

  const current = CUSTOMER_ROUTES.some((route) => isNavItemActive(pathname, route, "prefix"));
  const displayName = customer.profile.display_name?.trim();
  const name = accountName(displayName, customer.email);
  const next = glance?.next;
  const unread = glance?.unread ?? 0;

  return (
    // Not modal, as the dashboard's account menu: a short menu has no reason to
    // lock the page behind it.
    <DropdownMenu.Root modal={false}>
      <DropdownMenu.Trigger
        aria-label={`Account, ${name}`}
        className={cn("storefront-header-account", current && "storefront-header-account--current")}
      >
        <span aria-hidden="true" className="storefront-header-account__initial">
          {initialOf(displayName, customer.email)}
        </span>
        <span className="storefront-header-account__name">{name}</span>
        <ChevronDownIcon className="storefront-header-account__chevron" />
      </DropdownMenu.Trigger>

      <DropdownMenu.Portal>
        <DropdownMenu.Content
          align="end"
          className="dashboard-account-popover storefront-account-popover"
          sideOffset={8}
          onCloseAutoFocus={() => {
            if (signOutAsked.current) {
              signOutAsked.current = false;
              setSigningOut(true);
            }
          }}
        >
          <DropdownMenu.Label className="storefront-account-popover__who">
            {displayName ? <span className="storefront-account-popover__name">{displayName}</span> : null}
            <span title={customer.email}>{customer.email}</span>
          </DropdownMenu.Label>

          {next?.pickup_at ? (
            <DropdownMenu.Item asChild>
              <Link className="storefront-account-popover__next" href={ROUTE.orderDetail(next.id)}>
                <span className="storefront-account-popover__next-label">Next pickup</span>
                <span className="storefront-account-popover__next-time">{formatPickupDateTime(next.pickup_at)}</span>
                <span className="storefront-account-popover__next-code">{next.code}</span>
              </Link>
            </DropdownMenu.Item>
          ) : null}

          <DropdownMenu.Item asChild>
            <Link className="dashboard-account-item" href={ROUTE.orders}>
              Orders{" "}
              {unread > 0 ? <span className="storefront-account-popover__meta">{plural(unread, "new message")}</span> : null}
            </Link>
          </DropdownMenu.Item>
          <DropdownMenu.Item asChild>
            <Link className="dashboard-account-item" href={ROUTE.customer.saved}>
              Saved
            </Link>
          </DropdownMenu.Item>
          <DropdownMenu.Item asChild>
            <Link className="dashboard-account-item" href={ROUTE.customer.account.profile}>
              Account
            </Link>
          </DropdownMenu.Item>

          <DropdownMenu.Separator className="dashboard-account-separator" />

          <DropdownMenu.Item
            className="dashboard-account-item dashboard-account-item--signout"
            onSelect={() => {
              signOutAsked.current = true;
            }}
          >
            Sign out
            <LogOutIcon className="dashboard-account-item__mark" />
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
      <SignOutDialog open={signingOut} onClose={() => setSigningOut(false)} />
    </DropdownMenu.Root>
  );
}
