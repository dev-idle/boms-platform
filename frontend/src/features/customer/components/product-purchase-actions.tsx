"use client";

import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";
import type { ReactNode } from "react";

import { Button } from "@/components/ui/button";
import type { CatalogProduct } from "@/lib/schemas/catalog";
import {
  loginHrefPreservingNext,
  registerHrefPreservingNext,
} from "@/lib/validate-next";
import { useAuthStore } from "@/stores/auth-store";

import { AddToCartButton } from "./add-to-cart-button";
import { CustomCakeForm } from "./custom-cake-form";

type ProductPurchaseActionsProps = {
  product?: CatalogProduct;
  comboId?: string;
  /** On a card among others: an outline action, and a guest is only asked to sign in. */
  secondary?: boolean;
  /** On the product's page: what stands beside the action. */
  children?: ReactNode;
  /** On the product's page: told what a custom cake's chosen options add to its price. */
  onOptionsCentsChange?: (cents: number) => void;
};

export function ProductPurchaseActions({
  product,
  comboId,
  secondary = false,
  children,
  onOptionsCentsChange,
}: ProductPurchaseActionsProps) {
  const status = useAuthStore((state) => state.status);
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const search = searchParams.toString()
    ? `?${searchParams.toString()}`
    : "";

  // While a session is restored it is not known yet whether to offer the cart or
  // the sign-in, so the action waits in place instead of asking a customer to sign in.
  if (status === "idle") {
    return (
      <>
        <Button
          aria-busy
          disabled
          size={secondary ? "sm" : "default"}
          title="Checking your account"
          type="button"
          variant={secondary ? "outline" : "default"}
        >
          Add to cart
        </Button>
        {children}
      </>
    );
  }

  if (status === "authenticated") {
    if (product?.is_customizable) {
      return (
        <CustomCakeForm product={product} onOptionsCentsChange={onOptionsCentsChange}>
          {children}
        </CustomCakeForm>
      );
    }
    return (
      <>
        <AddToCartButton
          comboId={comboId}
          productId={product?.id}
          secondary={secondary}
        />
        {children}
      </>
    );
  }

  if (secondary) {
    return (
      <Button asChild size="sm" variant="outline">
        <Link href={loginHrefPreservingNext(pathname, search)}>Sign in to add</Link>
      </Button>
    );
  }

  return (
    <>
      <div className="storefront-guest-purchase">
        <Button asChild>
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
      {children}
    </>
  );
}
