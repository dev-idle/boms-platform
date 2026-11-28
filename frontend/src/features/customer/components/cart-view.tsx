"use client";

import Link from "next/link";
import { useState } from "react";

import { StorefrontEmptyState, StorefrontErrorState } from "@/components/layouts/storefront-state";
import { CloseIcon } from "@/components/icons/storefront-icons";
import { Button } from "@/components/ui/button";
import { CustomizationSummary } from "@/components/ui/customization-summary";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ROUTE } from "@/constants/routes";
import { CatalogTile } from "@/features/catalog";
import { formatPriceCents } from "@/lib/validation/catalog";

import {
  useApplyCartDiscount,
  useCart,
  useRemoveCartDiscount,
  useRemoveCartItem,
  useUpdateCartItem,
} from "../hooks";
import type { Cart, CartItem } from "../schemas";
import { CartCheckoutPanel } from "./cart-checkout-panel";
import { QuantityStepper } from "./quantity-stepper";

const DISCOUNT_INPUT_ID = "cart-discount-code";

function CartLineItem({ item }: { item: CartItem }) {
  const updateItem = useUpdateCartItem(item.id);
  const removeItem = useRemoveCartItem();
  const eyebrow = item.line_type === "combo" ? "Combo" : (item.category_name ?? "Product");

  return (
    <li className="storefront-cart-line">
      <CatalogTile name={item.name} scale="row" url={item.image_url} />
      <div className="storefront-cart-line__copy">
        <p className="storefront-line-eyebrow">{eyebrow}</p>
        {item.product_id ? (
          <Link className="storefront-line-name" href={ROUTE.productDetail(item.product_id)}>
            {item.name}
          </Link>
        ) : (
          <p className="storefront-line-name">{item.name}</p>
        )}
        <p className="storefront-line-meta">
          <span className="text-price">{formatPriceCents(item.unit_price_cents)}</span> each
        </p>
        {item.customization ? <CustomizationSummary customization={item.customization} /> : null}
        {!item.is_available ? (
          <p className="storefront-cart-line__warning text-caption">
            No longer available — remove it to check out.
          </p>
        ) : null}
      </div>
      <div className="storefront-cart-line__qty">
        <QuantityStepper
          itemName={item.name}
          lockedReason={item.is_available ? undefined : "This item is no longer available."}
          max={99}
          min={1}
          onChange={(quantity) => updateItem.mutate({ quantity })}
          pending={updateItem.isPending}
          value={item.quantity}
        />
      </div>
      <p className="storefront-cart-line__total text-price">{formatPriceCents(item.line_total_cents)}</p>
      <button
        aria-busy={removeItem.isPending || undefined}
        aria-label={`Remove ${item.name}`}
        className="storefront-line-remove"
        disabled={removeItem.isPending}
        onClick={() => removeItem.mutate(item.id)}
        title="Remove"
        type="button"
      >
        <CloseIcon />
      </button>
    </li>
  );
}

/** The code the customer enters, or the one applied with the way to take it off. */
function CartDiscountField({ cart }: { cart: Cart }) {
  const applyDiscount = useApplyCartDiscount();
  const removeDiscount = useRemoveCartDiscount();
  const [code, setCode] = useState("");

  if (cart.discount) {
    return (
      <div className="storefront-cart-summary__row">
        <span>
          Code <span className="storefront-cart-summary__code">{cart.discount.code}</span>
        </span>
        <button
          aria-busy={removeDiscount.isPending || undefined}
          className="storefront-text-action"
          disabled={removeDiscount.isPending}
          onClick={() => removeDiscount.mutate()}
          type="button"
        >
          {removeDiscount.isPending ? "Removing…" : "Remove"}
        </button>
      </div>
    );
  }

  return (
    <div className="storefront-cart-discount">
      <Label htmlFor={DISCOUNT_INPUT_ID}>Discount code</Label>
      <div className="storefront-cart-discount__controls">
        <Input
          id={DISCOUNT_INPUT_ID}
          onChange={(event) => setCode(event.target.value)}
          placeholder="Enter code"
          value={code}
        />
        <Button
          aria-busy={applyDiscount.isPending || undefined}
          disabled={applyDiscount.isPending || code.trim() === ""}
          onClick={() => applyDiscount.mutate({ code: code.trim() })}
          size="sm"
          title={code.trim() === "" ? "Enter a code first." : undefined}
          type="button"
          variant="outline"
        >
          {applyDiscount.isPending ? "Applying…" : "Apply"}
        </Button>
      </div>
    </div>
  );
}

/** The cart's own shape while it loads: a few rows beside the summary. */
function CartSkeleton() {
  return (
    <div aria-busy="true" className="storefront-detail-layout">
      <div className="grid gap-6">
        {[0, 1].map((row) => (
          <div className="flex gap-6" key={row}>
            <span className="catalog-tile catalog-tile--row skeleton" />
            <div className="grid flex-1 content-start gap-3">
              <span className="skeleton block h-3 w-24" />
              <span className="skeleton block h-6 w-2/3" />
              <span className="skeleton block h-4 w-28" />
            </div>
          </div>
        ))}
      </div>
      <span className="skeleton block h-80 w-full" />
    </div>
  );
}

export function CartView() {
  const cartQuery = useCart();

  if (cartQuery.isPending) {
    return <CartSkeleton />;
  }

  if (cartQuery.isError) {
    return (
      <StorefrontErrorState
        message="Check your connection and try again."
        onRetry={() => void cartQuery.refetch()}
        title="We could not load your cart"
      />
    );
  }

  const cart = cartQuery.data;

  if (cart.items.length === 0) {
    return (
      <StorefrontEmptyState
        action={
          <Button asChild>
            <Link href={ROUTE.products}>Browse all products</Link>
          </Button>
        }
        title="Your cart is empty"
      />
    );
  }

  return (
    <div className="storefront-detail-layout">
      <section aria-label="Items" className="storefront-cart__main">
        <div aria-hidden="true" className="storefront-cart-head">
          <span>Item</span>
          <span>Quantity</span>
          <span>Total</span>
        </div>
        <ul className="storefront-cart__lines">
          {cart.items.map((item) => (
            <CartLineItem key={item.id} item={item} />
          ))}
        </ul>
      </section>

      <aside aria-labelledby="cart-summary-title" className="storefront-detail-rail">
        <div className="storefront-cart-summary">
          <h2 className="storefront-cart-summary__title" id="cart-summary-title">
            Order summary
          </h2>
          <CartCheckoutPanel
            checkoutReady={cart.checkout_ready}
            fulfillment={cart.fulfillment}
            totalCents={cart.total_cents}
          >
            <dl className="storefront-cart-summary__rows">
              <div className="storefront-cart-summary__row">
                <dt>Subtotal</dt>
                <dd className="text-price">{formatPriceCents(cart.subtotal_cents)}</dd>
              </div>
              {cart.discount ? (
                <div className="storefront-cart-summary__row">
                  <dt>Discount</dt>
                  <dd className="text-price">−{formatPriceCents(cart.discount_cents)}</dd>
                </div>
              ) : null}
            </dl>
            <CartDiscountField cart={cart} />
            <div className="storefront-cart-summary__total">
              <span className="storefront-cart-summary__total-label">Total</span>
              <span className="text-price">{formatPriceCents(cart.total_cents)}</span>
            </div>
          </CartCheckoutPanel>
        </div>
      </aside>
    </div>
  );
}
