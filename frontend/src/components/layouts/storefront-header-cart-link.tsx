"use client";

import Link from "next/link";

import { CartIcon } from "@/components/icons/storefront-icons";
import { ROUTE } from "@/constants/routes";
import { useCart } from "@/features/customer";
import { useAuthStore } from "@/stores/auth-store";

function cartItemCount(items: { quantity: number }[]): number {
  return items.reduce((total, item) => total + item.quantity, 0);
}

/** Cart affordance with matcha count badge. */
export function StorefrontHeaderCartLink() {
  const isAuthenticated = useAuthStore((state) => state.status === "authenticated");
  const cartQuery = useCart({ enabled: isAuthenticated });
  const itemCount = cartQuery.data ? cartItemCount(cartQuery.data.items) : 0;

  return (
    <Link
      aria-busy={cartQuery.isPending || undefined}
      aria-label={
        itemCount > 0 ? `Cart, ${itemCount} items` : "Cart"
      }
      className="storefront-header-cart"
      href={ROUTE.cart}
    >
      <CartIcon />
      Cart
      {itemCount > 0 ? (
        <span className="storefront-header-cart__badge">{itemCount}</span>
      ) : null}
    </Link>
  );
}
