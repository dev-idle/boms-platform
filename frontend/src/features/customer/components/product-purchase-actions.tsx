"use client";

import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";

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
  label?: string;
};

export function ProductPurchaseActions({
  product,
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
    if (product?.is_customizable) {
      return <CustomCakeForm product={product} />;
    }
    return (
      <AddToCartButton
        comboId={comboId}
        label={label}
        productId={product?.id}
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
