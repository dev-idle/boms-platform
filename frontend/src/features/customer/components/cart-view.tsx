"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { formatPriceCents } from "@/lib/validation/catalog";

import {
  useApplyCartDiscount,
  useCart,
  useCheckoutCart,
  useRemoveCartDiscount,
  useRemoveCartItem,
  useUpdateCartItem,
} from "../hooks";
import type { CartItem } from "../schemas";

function CartLineItem({ item }: { item: CartItem }) {
  const updateItem = useUpdateCartItem(item.id);
  const removeItem = useRemoveCartItem();

  return (
    <li className="flex flex-col gap-3 rounded-lg border border-border p-4 rounded-card border bg-surface">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="font-medium text-ink">{item.name}</p>
          <p className="text-sm text-muted">
            {item.line_type === "combo" ? "Combo" : "Product"} ·{" "}
            {formatPriceCents(item.unit_price_cents)} each
          </p>
          {!item.is_available ? (
            <p className="mt-1 text-sm text-amber-700">
              No longer available — remove to continue checkout.
            </p>
          ) : null}
        </div>
        <p className="font-medium text-ink">
          {formatPriceCents(item.line_total_cents)}
        </p>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <label className="sr-only" htmlFor={`qty-${item.id}`}>
          Quantity for {item.name}
        </label>
        <Input
          className="w-20"
          disabled={updateItem.isPending || !item.is_available}
          variant="inline"
          id={`qty-${item.id}`}
          max={99}
          min={1}
          type="number"
          value={item.quantity}
          onChange={(event) => {
            const next = Number.parseInt(event.target.value, 10);
            if (Number.isNaN(next) || next < 1 || next > 99) {
              return;
            }
            updateItem.mutate({ quantity: next });
          }}
        />
        <Button
          disabled={removeItem.isPending}
          type="button"
          variant="outline"
          onClick={() => removeItem.mutate(item.id)}
        >
          Remove
        </Button>
      </div>
    </li>
  );
}

export function CartView() {
  const router = useRouter();
  const cartQuery = useCart();
  const applyDiscount = useApplyCartDiscount();
  const removeDiscount = useRemoveCartDiscount();
  const checkout = useCheckoutCart();
  const [discountCode, setDiscountCode] = useState("");

  if (cartQuery.isPending) {
    return <p className="text-sm text-muted">Loading cart…</p>;
  }

  if (cartQuery.isError) {
    const message =
      isApiError(cartQuery.error) && cartQuery.error.isAuthError()
        ? "Sign in to view your cart."
        : "Failed to load cart.";
    return <p className="text-sm text-error">{message}</p>;
  }

  const cart = cartQuery.data;
  if (!cart) {
    return null;
  }

  return (
    <div className="space-y-6">
      {cart.items.length === 0 ? (
        <div className="space-y-4">
          <p className="text-sm text-muted">
            Your cart is empty.
          </p>
          <Link href={ROUTE.products}>
            <Button type="button" variant="outline">
              Browse products
            </Button>
          </Link>
        </div>
      ) : (
        <>
          <ul className="space-y-3">
            {cart.items.map((item) => (
              <CartLineItem key={item.id} item={item} />
            ))}
          </ul>

          <div className="space-y-3 rounded-lg border border-border p-4 rounded-card border bg-surface">
            <div className="flex justify-between text-sm">
              <span className="text-muted">Subtotal</span>
              <span>{formatPriceCents(cart.subtotal_cents)}</span>
            </div>
            {cart.discount ? (
              <div className="flex justify-between text-sm text-emerald-700">
                <span>Discount ({cart.discount.code})</span>
                <span>-{formatPriceCents(cart.discount_cents)}</span>
              </div>
            ) : null}
            <div className="flex justify-between text-base font-medium">
              <span>Total</span>
              <span>{formatPriceCents(cart.total_cents)}</span>
            </div>
          </div>

          <div className="space-y-2">
            <p className="text-sm font-medium text-ink">
              Discount code
            </p>
            <div className="flex flex-wrap gap-2">
              <Input
                className="max-w-xs"
                placeholder="Enter code"
                value={discountCode}
                variant="inline"
                onChange={(event) => setDiscountCode(event.target.value)}
              />
              <Button
                disabled={applyDiscount.isPending || discountCode.trim() === ""}
                type="button"
                variant="outline"
                onClick={() =>
                  applyDiscount.mutate({ code: discountCode.trim() })
                }
              >
                Apply
              </Button>
              {cart.discount ? (
                <Button
                  disabled={removeDiscount.isPending}
                  type="button"
                  variant="ghost"
                  onClick={() => removeDiscount.mutate()}
                >
                  Remove discount
                </Button>
              ) : null}
            </div>
          </div>

          <Button
            disabled={!cart.checkout_ready || checkout.isPending}
            type="button"
            onClick={() =>
              checkout.mutate(undefined, {
                onSuccess: (order) => {
                  router.push(ROUTE.orderDetail(order.id));
                },
              })
            }
          >
            {checkout.isPending ? "Placing order…" : "Checkout"}
          </Button>
        </>
      )}
    </div>
  );
}
