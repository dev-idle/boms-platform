"use client";

import { useState, type ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { PolicyConsent } from "@/features/legal";
import { EmailVerificationNotice } from "@/features/user";
import { ApiErrorCode, isApiError } from "@/lib/errors";
import { formatPriceCents } from "@/lib/validation/catalog";
import { useAuthStore } from "@/stores/auth-store";

import { useCheckoutCart } from "../hooks";
import { usePickupChoice } from "../hooks/use-pickup-choice";
import type { Cart } from "../schemas";
import { PickupSlotPicker } from "./pickup-slot-picker";

const ERROR_ID = "pickup-error";
const CONSENT_HINT_ID = "checkout-consent-hint";

type CartCheckoutPanelProps = {
  checkoutReady: boolean;
  fulfillment: Cart["fulfillment"];
  totalCents: number;
  /** The money — totals, code, total — between the pickup and the checkout. */
  children: ReactNode;
};

/** Pickup day and slot, then the money, then checkout for the cart's items. */
export function CartCheckoutPanel({ checkoutReady, fulfillment, totalCents, children }: CartCheckoutPanelProps) {
  const checkout = useCheckoutCart();
  const choice = usePickupChoice(fulfillment, checkout.isPending);
  const [accepted, setAccepted] = useState(false);

  // Order updates go to the account's address, so it must be confirmed first.
  const unconfirmed = useAuthStore((state) => state.user?.email_verified === false);
  // Why the order cannot be placed yet, in the order the customer meets the causes.
  const blockedReason = !checkoutReady
    ? "Remove the items that are no longer available to place the order"
    : unconfirmed
      ? "Confirm your email to place an order"
      : !choice.ready
        ? "Choose a pickup time to place the order"
        : !accepted
          ? "Accept the terms of sale and the refund policy to place the order"
          : undefined;
  const canCheckout = blockedReason === undefined && !checkout.isPending;

  function placeOrder(): void {
    const pickupAt = canCheckout ? choice.pickupAt() : null;
    if (pickupAt === null) {
      return;
    }
    checkout.mutate(
      { pickup_at: pickupAt },
      {
        // The policies changed while the cart was open: the tick was given to an
        // older version, so it no longer counts.
        onError: (error) => {
          if (isApiError(error) && error.code === ApiErrorCode.TermsNotAccepted) {
            setAccepted(false);
          }
        },
      },
    );
  }

  return (
    <>
      <PickupSlotPicker {...choice.picker} errorId={ERROR_ID} />

      {children}

      <div className="storefront-cart-summary__checkout">
        <EmailVerificationNotice reason="Confirm your email to place this order." />
        <PolicyConsent
          checked={accepted}
          describedBy={accepted ? undefined : CONSENT_HINT_ID}
          disabled={checkout.isPending}
          onCheckedChange={setAccepted}
          scope="order"
        />
        {accepted ? null : (
          <p className="sr-only" id={CONSENT_HINT_ID}>
            Placing the order opens once you accept the terms of sale and the refund policy.
          </p>
        )}
        <Button
          className="storefront-cart-summary__checkout-btn"
          disabled={!canCheckout}
          title={blockedReason}
          type="button"
          onClick={placeOrder}
        >
          {checkout.isPending ? "Placing order…" : `Place order · ${formatPriceCents(totalCents)}`}
        </Button>
        <p className="storefront-pickup-picker__error text-caption" id={ERROR_ID} role="status">
          {choice.message}
        </p>
        {choice.retry ? (
          <Button size="sm" type="button" variant="outline" onClick={choice.retry}>
            Try again
          </Button>
        ) : null}
      </div>
    </>
  );
}
