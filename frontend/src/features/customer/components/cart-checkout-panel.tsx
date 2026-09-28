"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { PICKUP_COPY } from "@/constants/pickup";
import { ROUTE } from "@/constants/routes";
import { useNow } from "@/lib/hooks/use-now";
import {
  bakeryPickupISOFromLocalInput,
  defaultPickupLocalInputValue,
  maxPickupLocalInputValue,
  pickupProblem,
  type PickupWindow,
} from "@/lib/validation/pickup";

import { useCheckoutCart, usePickupRules } from "../hooks";
import { pickupHint, pickupProblemMessage, toPickupWindow } from "../lib/pickup-window";
import { PickupSlotPicker } from "./pickup-slot-picker";

const ERROR_ID = "pickup-at-error";
/** How often the suggested pickup and its checks move with the clock. */
const CLOCK_TICK_MS = 30_000;

type CartCheckoutPanelProps = {
  checkoutReady: boolean;
};

function pickupMessage(
  value: string,
  pickupWindow: PickupWindow,
  chosen: boolean,
  now: Date,
): string | null {
  const problem = pickupProblem(value, pickupWindow, now);
  if (problem === null) {
    return null;
  }
  if (problem === "missing") {
    return chosen ? PICKUP_COPY.required : PICKUP_COPY.noOpenTime;
  }
  return pickupProblemMessage(problem, pickupWindow, pickupWindow.closedDates.get(value.slice(0, 10)));
}

/**
 * Pickup time and checkout. The pickup rules are read from the API and refresh
 * live when an admin edits them; until the customer picks a time, the earliest
 * valid one follows those rules and the clock.
 */
export function CartCheckoutPanel({ checkoutReady }: CartCheckoutPanelProps) {
  const router = useRouter();
  const checkout = useCheckoutCart();
  const rulesQuery = usePickupRules();
  const now = useNow(CLOCK_TICK_MS);
  const [chosen, setChosen] = useState<string | null>(null);

  const pickupWindow = rulesQuery.data ? toPickupWindow(rulesQuery.data) : null;
  const earliest = pickupWindow ? defaultPickupLocalInputValue(pickupWindow, now) : "";
  const value = chosen ?? earliest;
  const message = pickupWindow
    ? pickupMessage(value, pickupWindow, chosen !== null, now)
    : rulesQuery.isError
      ? PICKUP_COPY.unavailable
      : null;
  const canCheckout =
    checkoutReady && pickupWindow !== null && message === null && !checkout.isPending;

  function placeOrder(): void {
    if (!canCheckout || !pickupWindow) {
      return;
    }
    // The suggestion on screen can be up to a tick old; take a fresh one.
    const at = new Date();
    const pickup = chosen ?? defaultPickupLocalInputValue(pickupWindow, at);
    if (pickupProblem(pickup, pickupWindow, at) !== null) {
      return;
    }
    checkout.mutate(
      { pickup_at: bakeryPickupISOFromLocalInput(pickup) },
      { onSuccess: (order) => router.push(ROUTE.orderDetail(order.id)) },
    );
  }

  return (
    <>
      <PickupSlotPicker
        disabled={pickupWindow === null || checkout.isPending}
        errorId={ERROR_ID}
        hint={
          pickupWindow ? pickupHint(pickupWindow) : rulesQuery.isError ? "" : PICKUP_COPY.loading
        }
        invalid={message !== null}
        max={pickupWindow ? maxPickupLocalInputValue(pickupWindow, now) : ""}
        min={earliest}
        value={value}
        onChange={setChosen}
      />

      <div className="storefront-cart-summary__checkout">
        <Button
          className="storefront-cart-summary__checkout-btn"
          disabled={!canCheckout}
          type="button"
          onClick={placeOrder}
        >
          {checkout.isPending ? "Placing order…" : "Checkout"}
        </Button>
        <p className="storefront-cart-summary__pickup-error text-caption" id={ERROR_ID} role="status">
          {message}
        </p>
        {rulesQuery.isError ? (
          <Button type="button" variant="outline" onClick={() => void rulesQuery.refetch()}>
            Try again
          </Button>
        ) : null}
      </div>
    </>
  );
}
