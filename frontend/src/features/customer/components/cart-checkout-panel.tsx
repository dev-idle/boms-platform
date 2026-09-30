"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { PICKUP_COPY } from "@/constants/pickup";
import { ROUTE } from "@/constants/routes";
import { PolicyConsent } from "@/features/legal";
import { EmailVerificationNotice } from "@/features/user";
import { ApiErrorCode, isApiError } from "@/lib/errors";
import { useNow } from "@/lib/hooks/use-now";
import { clockToMinutes, formatClockMinutes } from "@/lib/validation/clock";
import { useAuthStore } from "@/stores/auth-store";
import {
  bakeryDayOf,
  bakeryPickupISOFromLocalInput,
  earliestPickupLocalValue,
  formatPickupLocalInputValue,
  lastPickupDay,
  pickupOrderType,
  pickupProblem,
  type PickupFulfillment,
  type PickupWindow,
} from "@/lib/validation/pickup";

import { useCheckoutCart, usePickupRules, usePickupSlots } from "../hooks";
import {
  orderTypeNote,
  pickupHint,
  pickupProblemMessage,
  toPickupFulfillment,
  toPickupWindow,
} from "../lib/pickup-window";
import type { Cart, PickupSlots } from "../schemas";
import { PickupSlotPicker, type PickupSlotOption } from "./pickup-slot-picker";

const ERROR_ID = "pickup-error";
const CONSENT_HINT_ID = "checkout-consent-hint";
/** How often the suggested pickup and its checks move with the clock. */
const CLOCK_TICK_MS = 30_000;

type CartCheckoutPanelProps = {
  checkoutReady: boolean;
  fulfillment: Cart["fulfillment"];
};

/**
 * The day's slots still ahead, as options: a full slot, or one too soon for
 * these items, stays in the list, greyed with the reason.
 */
function slotOptions(
  slots: PickupSlots["slots"],
  pickupWindow: PickupWindow,
  now: Date,
  items: PickupFulfillment,
): PickupSlotOption[] {
  return slots
    .filter(({ starts_at }) => new Date(starts_at).getTime() > now.getTime())
    .map(({ starts_at, full }) => {
      const value = formatPickupLocalInputValue(new Date(starts_at));
      const time = formatClockMinutes(clockToMinutes(value.slice(11)));
      const tooSoon = pickupProblem(value, pickupWindow, now, items) !== null;
      const reason = full ? PICKUP_COPY.full : tooSoon ? PICKUP_COPY.tooSoon : null;
      return { value, label: reason ? `${time} · ${reason}` : time, full, disabled: full || tooSoon };
    });
}

/**
 * Pickup day, slot and checkout. The rules are read from the API and refresh
 * live when an admin edits them, the day's slots as they fill and free up;
 * until the customer chooses, the earliest slot that suits the cart's items is
 * offered.
 */
export function CartCheckoutPanel({ checkoutReady, fulfillment }: CartCheckoutPanelProps) {
  const router = useRouter();
  const checkout = useCheckoutCart();
  const rulesQuery = usePickupRules();
  const now = useNow(CLOCK_TICK_MS);
  const [chosenDay, setChosenDay] = useState<string | null>(null);
  const [chosenSlot, setChosenSlot] = useState<string | null>(null);
  const [accepted, setAccepted] = useState(false);

  const items = toPickupFulfillment(fulfillment);
  const pickupWindow = rulesQuery.data ? toPickupWindow(rulesQuery.data) : null;
  const earliest = pickupWindow ? earliestPickupLocalValue(pickupWindow, now, items) : "";
  const minDay = earliest.slice(0, 10) || bakeryDayOf(now);
  const maxDay = pickupWindow ? lastPickupDay(pickupWindow, now) : "";
  const day = chosenDay ?? earliest.slice(0, 10);
  const closedReason = pickupWindow?.closedDates.get(day);

  function dayProblem(): string | null {
    if (!pickupWindow) {
      return rulesQuery.isError ? PICKUP_COPY.unavailable : null;
    }
    if (earliest === "") {
      return PICKUP_COPY.noOpenTime;
    }
    if (day === "") {
      return PICKUP_COPY.required;
    }
    if (day < minDay) {
      return pickupProblemMessage("too_soon", pickupWindow, items, pickupOrderType(earliest, now, items) ?? "pre_order");
    }
    if (day > maxDay) {
      return pickupProblemMessage("too_far", pickupWindow, items, "pre_order");
    }
    if (closedReason !== undefined) {
      return pickupProblemMessage("closed_day", pickupWindow, items, "pre_order", closedReason);
    }
    return null;
  }

  const dayMessage = dayProblem();
  const slotsQuery = usePickupSlots(pickupWindow && dayMessage === null ? day : "");
  const loadingSlots = dayMessage === null && slotsQuery.isFetching && slotsQuery.data?.date !== day;
  const slots =
    pickupWindow && dayMessage === null && slotsQuery.data?.date === day
      ? slotOptions(slotsQuery.data.slots, pickupWindow, now, items)
      : [];
  const slot = chosenSlot ?? slots.find((option) => !option.disabled)?.value ?? "";

  function slotProblem(): string | null {
    if (!pickupWindow || dayMessage !== null || loadingSlots) {
      return null;
    }
    if (slotsQuery.isError) {
      return PICKUP_COPY.slotsUnavailable;
    }
    if (slot === "") {
      return slotsQuery.isSuccess ? PICKUP_COPY.noSlotLeft : null;
    }
    if (slots.some((option) => option.value === slot && option.full)) {
      return PICKUP_COPY.slotFull;
    }
    const problem = pickupProblem(slot, pickupWindow, now, items);
    if (problem === null) {
      return null;
    }
    if (problem === "missing") {
      return PICKUP_COPY.required;
    }
    return pickupProblemMessage(problem, pickupWindow, items, pickupOrderType(slot, now, items) ?? "pre_order");
  }

  const slotMessage = slotProblem();
  const message = dayMessage ?? slotMessage;
  const type = pickupWindow && slot !== "" && message === null ? pickupOrderType(slot, now, items) : null;
  // Order updates go to the account's address, so it must be confirmed first.
  const unconfirmed = useAuthStore((state) => state.user?.email_verified === false);
  const canCheckout =
    checkoutReady &&
    !unconfirmed &&
    accepted &&
    pickupWindow !== null &&
    slot !== "" &&
    message === null &&
    !loadingSlots &&
    !checkout.isPending;

  function placeOrder(): void {
    if (!canCheckout || !pickupWindow) {
      return;
    }
    // The clock on screen can be up to a tick old; check the slot against now.
    if (pickupProblem(slot, pickupWindow, new Date(), items) !== null) {
      return;
    }
    checkout.mutate(
      { pickup_at: bakeryPickupISOFromLocalInput(slot) },
      {
        onSuccess: (order) => router.push(ROUTE.orderDetail(order.id)),
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
      <PickupSlotPicker
        day={day}
        dayInvalid={dayMessage !== null && pickupWindow !== null}
        disabled={pickupWindow === null || checkout.isPending}
        errorId={ERROR_ID}
        hint={pickupWindow ? pickupHint(pickupWindow) : rulesQuery.isError ? "" : PICKUP_COPY.loading}
        loadingSlots={loadingSlots}
        maxDay={maxDay}
        minDay={minDay}
        slot={slot}
        slotInvalid={slotMessage !== null}
        slots={slots}
        typeNote={pickupWindow && type ? orderTypeNote(type, pickupWindow, items) : ""}
        onDayChange={(next) => {
          setChosenDay(next);
          setChosenSlot(null);
        }}
        onSlotChange={setChosenSlot}
      />

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
        <p className="storefront-cart-summary__pickup-error text-caption" id={ERROR_ID} role="status">
          {message}
        </p>
        {rulesQuery.isError || (dayMessage === null && slotsQuery.isError) ? (
          <Button
            type="button"
            variant="outline"
            onClick={() => void (rulesQuery.isError ? rulesQuery.refetch() : slotsQuery.refetch())}
          >
            Try again
          </Button>
        ) : null}
      </div>
    </>
  );
}
