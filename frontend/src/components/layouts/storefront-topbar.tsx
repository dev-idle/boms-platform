"use client";

import Link from "next/link";

import { BRAND } from "@/constants/brand";
import { USER_ROLE } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";
import { StoreHours, useOrdersGlance } from "@/features/customer";
import { formatPickupDateTime } from "@/lib/validation/pickup";
import { useAuthStore } from "@/stores/auth-store";

/**
 * Utility strip above the sticky header — scrolls away with the page. A
 * signed-in customer with an order to collect sees when; everyone else sees
 * the opening hours and where to collect.
 */
export function StorefrontTopbar() {
  const signedIn = useAuthStore(
    (state) => state.status === "authenticated" && state.user?.role === USER_ROLE.customer,
  );
  const next = useOrdersGlance(signedIn).data?.next;

  return (
    <div className="storefront-topbar">
      <p className="storefront-container storefront-topbar__inner">
        {signedIn && next?.pickup_at ? (
          <>
            <span>
              Next pickup <strong>{formatPickupDateTime(next.pickup_at)}</strong>
            </span>
            <Link className="storefront-topbar__link" href={ROUTE.orderDetail(next.id)}>
              View order
            </Link>
          </>
        ) : (
          <span>
            Open daily <StoreHours /> · Pickup only in {BRAND.addressLine}
          </span>
        )}
      </p>
    </div>
  );
}
