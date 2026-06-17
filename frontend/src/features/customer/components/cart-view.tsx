"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { ROUTE } from "@/constants/routes";
import { STOREFRONT_NAV_COPY } from "@/constants/storefront-nav-copy";
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
    <li className="storefront-cart-line">
      <div className="storefront-cart-line__header">
        <div className="storefront-cart-line__copy">
          <p className="storefront-cart-line__name">{item.name}</p>
          <p className="storefront-cart-line__meta text-caption">
            {item.line_type === "combo" ? "Combo" : "Product"} ·{" "}
            {formatPriceCents(item.unit_price_cents)} each
          </p>
          {!item.is_available ? (
            <p className="storefront-cart-line__warning text-caption">
              No longer available — remove to continue checkout.
            </p>
          ) : null}
        </div>
        <p className="text-price storefront-cart-line__total">
          {formatPriceCents(item.line_total_cents)}
        </p>
      </div>
      <div className="storefront-cart-line__actions">
        <label className="sr-only" htmlFor={`qty-${item.id}`}>
          Quantity for {item.name}
        </label>
        <Input
          className="storefront-cart-line__qty"
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
        <button
          className="storefront-cart-remove"
          disabled={removeItem.isPending}
          type="button"
          onClick={() => removeItem.mutate(item.id)}
        >
          <span aria-hidden="true" className="storefront-cart-remove__icon">
            ×
          </span>
          <span>Remove</span>
        </button>
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
    return <InlineLoadingState className="storefront-customer-loading" />;
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
    <div className="storefront-cart">
      {cart.items.length === 0 ? (
        <div className="storefront-empty-state storefront-empty-state--card">
          <p className="storefront-empty-state__message">Your cart is empty.</p>
          <Button asChild variant="outline">
            <Link href={ROUTE.products}>{STOREFRONT_NAV_COPY.returnToShop}</Link>
          </Button>
        </div>
      ) : (
        <div className="storefront-cart-layout">
          <div className="storefront-cart__main">
            <ul className="storefront-cart__lines">
              {cart.items.map((item) => (
                <CartLineItem key={item.id} item={item} />
              ))}
            </ul>

            <div className="storefront-panel storefront-cart-discount">
              <p className="storefront-cart-discount__label">Discount code</p>
              <div className="storefront-cart-discount__controls">
                <Input
                  className="storefront-cart-discount__input"
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
                  <button
                    className="storefront-cart-remove storefront-cart-remove--code"
                    disabled={removeDiscount.isPending}
                    type="button"
                    onClick={() => removeDiscount.mutate()}
                  >
                    <span aria-hidden="true" className="storefront-cart-remove__icon">
                      ×
                    </span>
                    <span>Remove code</span>
                  </button>
                ) : null}
              </div>
            </div>
          </div>

          <aside className="storefront-cart__aside">
            <div className="storefront-panel storefront-cart-summary">
              <p className="storefront-cart-summary__title">Order summary</p>
              <div className="storefront-cart-summary__row">
                <span className="text-muted">Subtotal</span>
                <span className="text-table-cell">
                  {formatPriceCents(cart.subtotal_cents)}
                </span>
              </div>
              {cart.discount ? (
                <div className="storefront-cart-summary__row storefront-cart-summary__row--discount">
                  <span>Discount ({cart.discount.code})</span>
                  <span>-{formatPriceCents(cart.discount_cents)}</span>
                </div>
              ) : null}
              <div className="storefront-cart-summary__row storefront-cart-summary__row--total">
                <span>Total</span>
                <span className="text-price">
                  {formatPriceCents(cart.total_cents)}
                </span>
              </div>

              <div className="storefront-cart-summary__checkout">
                <Button
                  className="storefront-cart-summary__checkout-btn"
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
              </div>
            </div>
          </aside>
        </div>
      )}
    </div>
  );
}
