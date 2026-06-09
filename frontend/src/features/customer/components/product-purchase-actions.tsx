"use client";

import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";

import { Button } from "@/components/ui/button";
import { loginHrefPreservingNext } from "@/lib/validate-next";
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
  const query = searchParams.toString();
  const returnPath = query ? `${pathname}?${query}` : pathname;

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
    <Button asChild size="lg">
      <Link href={loginHrefPreservingNext(returnPath)}>
        Sign in to order
      </Link>
    </Button>
  );
}
