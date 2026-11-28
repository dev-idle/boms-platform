"use client";

import Link from "next/link";
import { useId, useRef, useState } from "react";

import { CartIcon } from "@/components/icons/storefront-icons";
import { USER_ROLE } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";
import { cartItemCount, MiniCart, useCart } from "@/features/customer";
import { cn, plural } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

import { useStorefrontHeaderPanel } from "./use-storefront-header-panel";

/**
 * The cart in the header: a signed-in customer drops it open under the icon,
 * which fills in matcha with the count beside the bag once it holds something;
 * a guest goes to the cart page, which asks them to sign in.
 */
export function StorefrontHeaderCartLink() {
  const [open, setOpen] = useState(false);
  const panelId = useId();
  const panelRef = useRef<HTMLDivElement>(null);
  const toggleRef = useRef<HTMLButtonElement>(null);
  const signedIn = useAuthStore(
    (state) => state.status === "authenticated" && state.user?.role === USER_ROLE.customer,
  );
  const cartQuery = useCart({ enabled: signedIn });
  const itemCount = cartQuery.data ? cartItemCount(cartQuery.data.items) : 0;
  const label = itemCount > 0 ? `Cart, ${plural(itemCount, "item")}` : "Cart";

  // Closed from inside, the focus goes back to the icon instead of the page.
  function close(): void {
    if (panelRef.current?.contains(document.activeElement)) {
      toggleRef.current?.focus();
    }
    setOpen(false);
  }

  useStorefrontHeaderPanel({ open, onClose: close, panelRef, toggleRef, focusRef: panelRef });

  if (!signedIn) {
    return (
      <Link aria-label={label} className="storefront-header-cart" href={ROUTE.cart}>
        <CartIcon />
      </Link>
    );
  }

  return (
    <>
      <button
        aria-busy={cartQuery.isPending || undefined}
        aria-controls={open ? panelId : undefined}
        aria-expanded={open}
        aria-haspopup="dialog"
        aria-label={label}
        className={cn("storefront-header-cart", itemCount > 0 && "storefront-header-cart--filled")}
        onClick={() => setOpen((current) => !current)}
        ref={toggleRef}
        type="button"
      >
        <CartIcon />
        {itemCount > 0 ? <span className="storefront-header-cart__count">{itemCount}</span> : null}
      </button>
      {open ? <MiniCart id={panelId} onNavigate={() => setOpen(false)} panelRef={panelRef} /> : null}
    </>
  );
}
