"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { PolicyConsent } from "@/features/legal";
import { EmailVerificationNotice } from "@/features/user";
import { ApiErrorCode, isApiError } from "@/lib/errors";
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
};

/** Pickup day, slot and checkout for the cart's items. */
export function CartCheckoutPanel({ checkoutReady, fulfillment }: CartCheckoutPanelProps) {
  const checkout = useCheckoutCart();
  const choice = usePickupChoice(fulfillment, checkout.isPending);
  const [accepted, setAccepted] = useState(false);

  // Order updates go to the account's address, so it must be confirmed first.
  const unconfirmed = useAuthStore((state) => state.user?.email_verified === false);
  const canCheckout = checkoutReady && !unconfirmed && accepted && choice.ready && !checkout.isPending;

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

      <div className="storefront-cart-summary__checkout">
        <EmailVerificationNotice reason="Confirm your address to place this order." />
        <PolicyConsent
          checked={accepted}
          describedBy={accepted ? undefined : CONSENT_HINT_ID}
          disabled={checkout.isPending}
          onCheckedChange={setAccepted}
          scope="order"
        />
        {accepted ? null : (
          <p className="sr-only" id={CONSENT_HINT_ID}>
            Checkout opens once you accept the terms and the refund policy.
          </p>
        )}
        <Button
          className="storefront-cart-summary__checkout-btn"
          disabled={!canCheckout}
          title={
            unconfirmed
              ? "Confirm your email to place an order"
              : accepted
                ? undefined
                : "Accept the terms and the refund policy to check out"
          }
          type="button"
          onClick={placeOrder}
        >
          {checkout.isPending ? "Placing order…" : "Checkout"}
        </Button>
        <p className="storefront-pickup-picker__error text-caption" id={ERROR_ID} role="status">
          {choice.message}
        </p>
        {choice.retry ? (
          <Button type="button" variant="outline" onClick={choice.retry}>
            Try again
          </Button>
        ) : null}
      </div>
    </>
  );
}
