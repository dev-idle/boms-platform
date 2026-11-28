"use client";

import Link from "next/link";
import { useId, type RefObject } from "react";

import { StorefrontEmptyState, StorefrontErrorState } from "@/components/layouts/storefront-state";
import { Button } from "@/components/ui/button";
import { CustomizationSummary } from "@/components/ui/customization-summary";
import { ROUTE } from "@/constants/routes";
import { CatalogTile } from "@/features/catalog";
import { plural } from "@/lib/utils";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useCart, useRemoveCartItem } from "../hooks";
import { cartItemCount } from "../lib/cart-count";
import type { CartItem } from "../schemas";

type MiniCartProps = {
  id: string;
  /** Leaves the cart for another page. */
  onNavigate: () => void;
  panelRef: RefObject<HTMLDivElement | null>;
};

type MiniCartLineProps = {
  item: CartItem;
  /** Called once the line is gone, so focus can move off the button that went with it. */
  onRemoved: () => void;
};

function MiniCartLine({ item, onRemoved }: MiniCartLineProps) {
  const removeItem = useRemoveCartItem();

  return (
    <li className="storefront-mini-cart__line">
      <CatalogTile name={item.name} scale="row" url={item.image_url} />
      <div className="storefront-mini-cart__copy">
        <p className="storefront-mini-cart__head">
          <span className="storefront-mini-cart__name">{item.name}</span>
          <span className="text-price">{formatPriceCents(item.line_total_cents)}</span>
        </p>
        {item.customization ? <CustomizationSummary customization={item.customization} /> : null}
        {!item.is_available ? (
          <p className="storefront-cart-line__warning text-caption">
            No longer available — remove it to check out.
          </p>
        ) : null}
        <div className="storefront-mini-cart__meta">
          <span>
            {item.quantity} × {formatPriceCents(item.unit_price_cents)}
          </span>
          <button
            aria-label={`Remove ${item.name}`}
            className="storefront-text-action"
            disabled={removeItem.isPending}
            onClick={() => removeItem.mutate(item.id, { onSuccess: onRemoved })}
            type="button"
          >
            Remove
          </button>
        </div>
      </div>
    </li>
  );
}

/** The cart dropped from the header: its lines, the subtotal and the way to checkout. */
export function MiniCart({ id, onNavigate, panelRef }: MiniCartProps) {
  const cartQuery = useCart();
  const titleId = useId();
  const cart = cartQuery.data;
  const itemCount = cart ? cartItemCount(cart.items) : 0;

  return (
    <div aria-labelledby={titleId} className="storefront-mini-cart" id={id} ref={panelRef} role="dialog" tabIndex={-1}>
      <div className="storefront-mini-cart__title">
        <h2 className="storefront-mini-cart__heading" id={titleId}>
          Cart
        </h2>
        {itemCount > 0 ? <span className="storefront-mini-cart__count">{plural(itemCount, "item")}</span> : null}
      </div>
      {cartQuery.isPending ? (
        <ul aria-busy="true" aria-label="Loading your cart" className="storefront-mini-cart__lines">
          {[0, 1].map((row) => (
            <li className="storefront-mini-cart__line" key={row}>
              <span className="skeleton block size-16" />
              <span className="grid content-start gap-2">
                <span className="skeleton block h-4 w-3/4" />
                <span className="skeleton block h-3 w-1/3" />
              </span>
            </li>
          ))}
        </ul>
      ) : cartQuery.isError ? (
        <StorefrontErrorState onRetry={() => void cartQuery.refetch()} title="We could not load your cart" />
      ) : !cart || itemCount === 0 ? (
        <StorefrontEmptyState
          action={
            <Button asChild variant="outline">
              <Link href={ROUTE.products} onClick={onNavigate}>
                Browse all products
              </Link>
            </Button>
          }
          title="Your cart is empty"
        />
      ) : (
        <>
          <ul className="storefront-mini-cart__lines">
            {cart.items.map((item) => (
              <MiniCartLine item={item} key={item.id} onRemoved={() => panelRef.current?.focus()} />
            ))}
          </ul>
          <div className="storefront-mini-cart__footer">
            <p className="storefront-mini-cart__subtotal">
              <span>Subtotal</span>
              <span className="text-price">{formatPriceCents(cart.subtotal_cents)}</span>
            </p>
            <Button asChild className="w-full">
              <Link href={ROUTE.cart} onClick={onNavigate}>
                Checkout
              </Link>
            </Button>
          </div>
        </>
      )}
    </div>
  );
}
