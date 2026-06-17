"use client";

import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";

import { Button } from "@/components/ui/button";
import {
  loginHrefPreservingNext,
  registerHrefPreservingNext,
} from "@/lib/validate-next";
import { useAuthStore } from "@/stores/auth-store";

import { AddToCartButton } from "./add-to-cart-button";

type ProductPurchaseActionsProps = {
  productId?: string;
  comboId?: string;
  label?: string;
};

export function ProductPurchaseActions({
  productId,
  comboId,
  label = "Add to cart",
}: ProductPurchaseActionsProps) {
  const status = useAuthStore((state) => state.status);
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const search = searchParams.toString()
    ? `?${searchParams.toString()}`
    : "";

  if (status === "authenticated") {
    return (
      <AddToCartButton
        comboId={comboId}
        label={label}
        productId={productId}
      />
    );
  }

  return (
    <div className="storefront-guest-purchase">
      <Button asChild variant="outline">
        <Link href={loginHrefPreservingNext(pathname, search)}>
          Sign in to add to cart
        </Link>
      </Button>
      <p className="storefront-guest-purchase__hint text-caption">
        New here?{" "}
        <Link
          className="storefront-inline-link"
          href={registerHrefPreservingNext(pathname, search)}
        >
          Create an account
        </Link>
      </p>
    </div>
  );
}
