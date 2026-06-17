"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useCallback, useEffect, useId, useRef, useState } from "react";

import {
  OrdersIcon,
  SignOutIcon,
  UserIcon,
} from "@/components/icons/storefront-icons";
import { ROUTE } from "@/constants/routes";
import { isStorefrontCustomerNavActive } from "@/constants/storefront-customer-nav";
import { useLogout } from "@/features/auth/hooks";
import { cn } from "@/lib/utils";

import { StorefrontIconButton } from "./storefront-icon-button";

const ACCOUNT_LINKS = [
  { href: ROUTE.orders, label: "Orders", Icon: OrdersIcon },
  { href: ROUTE.customer.account.profile, label: "Account", Icon: UserIcon },
] as const;

/** Authenticated customer account menu — shop links + sign out. */
export function StorefrontAccountMenu() {
  const menuId = useId();
  const rootRef = useRef<HTMLDivElement>(null);
  const toggleRef = useRef<HTMLButtonElement>(null);
  const pathname = usePathname();
  const logout = useLogout();
  const [open, setOpen] = useState(false);

  const closeMenu = useCallback(() => setOpen(false), []);

  useEffect(() => {
    if (!open) {
      return;
    }

    function handlePointerDown(event: MouseEvent): void {
      if (!rootRef.current?.contains(event.target as Node)) {
        closeMenu();
      }
    }

    function handleKeyDown(event: KeyboardEvent): void {
      if (event.key === "Escape") {
        closeMenu();
        toggleRef.current?.focus();
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [closeMenu, open]);

  return (
    <div className="storefront-account-menu" ref={rootRef}>
      <StorefrontIconButton
        ref={toggleRef}
        aria-controls={open ? menuId : undefined}
        label="Account menu"
        onClick={() => setOpen((current) => !current)}
        pressed={open}
      >
        <UserIcon />
      </StorefrontIconButton>

      {open ? (
        <div
          className="storefront-account-menu__panel"
          id={menuId}
          role="menu"
        >
          <ul className="storefront-account-menu__list">
            {ACCOUNT_LINKS.map((item) => {
              const Icon = item.Icon;
              const active = isStorefrontCustomerNavActive(pathname, item.href);

              return (
                <li key={item.href} role="none">
                  <Link
                    className={cn(
                      "storefront-account-menu__item",
                      active && "storefront-account-menu__item--active",
                    )}
                    href={item.href}
                    onClick={closeMenu}
                    role="menuitem"
                  >
                    <Icon className="storefront-account-menu__icon" />
                    <span>{item.label}</span>
                  </Link>
                </li>
              );
            })}
          </ul>

          <div className="storefront-account-menu__footer">
            <button
              className="storefront-account-menu__sign-out"
              disabled={logout.isPending}
              onClick={() => {
                closeMenu();
                logout.mutate();
              }}
              role="menuitem"
              type="button"
            >
              <SignOutIcon className="storefront-account-menu__icon storefront-account-menu__icon--sign-out" />
              <span>{logout.isPending ? "Signing out…" : "Sign out"}</span>
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
