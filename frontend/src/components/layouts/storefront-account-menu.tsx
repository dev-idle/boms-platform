"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useCallback, useEffect, useId, useRef, useState } from "react";
import { UserIcon } from "@/components/icons/storefront-icons";
import { ROUTE } from "@/constants/routes";
import { useLogout } from "@/features/auth/hooks";
import { cn } from "@/lib/utils";

import { StorefrontIconButton } from "./storefront-icon-button";

const ACCOUNT_LINKS = [
  { href: ROUTE.customer.account.profile, label: "Profile" },
  { href: ROUTE.customer.account.password, label: "Password" },
  { href: ROUTE.orders, label: "Orders" },
] as const;

/** Authenticated customer account menu — profile links + sign out. */
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
            {ACCOUNT_LINKS.map((item) => (
              <li key={item.href} role="none">
                <Link
                  className={cn(
                    "storefront-account-menu__item",
                    pathname === item.href &&
                      "storefront-account-menu__item--active",
                  )}
                  href={item.href}
                  onClick={closeMenu}
                  role="menuitem"
                >
                  {item.label}
                </Link>
              </li>
            ))}
          </ul>

          <div className="storefront-account-menu__divider" role="separator" />

          <button
            className="storefront-account-menu__sign-out"
            disabled={logout.isPending}
            onClick={() => logout.mutate()}
            role="menuitem"
            type="button"
          >
            {logout.isPending ? "Signing out…" : "Sign out"}
          </button>
        </div>
      ) : null}
    </div>
  );
}
