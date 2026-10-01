"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { Button } from "@/components/ui/button";
import { PickupSlotPicker, usePickupChoice } from "@/features/customer";
import { isApiError } from "@/lib/errors";
import type { Fulfillment } from "@/lib/schemas/order";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useCreateStaffOrder, useStaffCustomer, useStaffOrderQuote } from "../hooks";
import { staffGuestFormSchema, type StaffGuestForm, type StaffOrderItemInput } from "../schemas";
import {
  StaffOrderCustomer,
  type StaffOrderChannel,
  type StaffOrderCustomerMode,
} from "./staff-order-customer";
import { StaffOrderItemsPicker, type StaffOrderLine } from "./staff-order-items-picker";

const PICKUP_ERROR_ID = "staff-order-pickup-error";

/** Until items are chosen there is nothing to ask of the bakery. */
const NO_ITEMS: Fulfillment = { has_kitchen_items: false, lead_minutes: 0, sold_out_on: null };

function toItemInput(line: StaffOrderLine): StaffOrderItemInput {
  return line.kind === "product"
    ? { product_id: line.id, quantity: line.quantity }
    : { combo_id: line.id, quantity: line.quantity };
}

/** An order taken at the counter or on the phone, paid in cash when it is collected. */
export function StaffNewOrder() {
  const [channel, setChannel] = useState<StaffOrderChannel>("counter");
  const [mode, setMode] = useState<StaffOrderCustomerMode>("guest");
  const [lookedUp, setLookedUp] = useState("");
  const [lines, setLines] = useState<StaffOrderLine[]>([]);
  const [problem, setProblem] = useState<string | null>(null);
  const items = lines.map(toItemInput);
  const customerQuery = useStaffCustomer(lookedUp);
  const quoteQuery = useStaffOrderQuote(items);
  const create = useCreateStaffOrder();
  const choice = usePickupChoice(quoteQuery.data?.fulfillment ?? NO_ITEMS, create.isPending);
  const guestForm = useForm<StaffGuestForm>({
    resolver: zodResolver(staffGuestFormSchema),
    defaultValues: { name: "", phone: "" },
  });
  // The total of the items before the last change, while their new total loads.
  const pricing = quoteQuery.isPlaceholderData || quoteQuery.isFetching;
  const priced = lines.length > 0 && quoteQuery.data !== undefined;

  function takeOrder(customer: { customer_id: string } | { guest: StaffGuestForm }): void {
    const pickupAt = choice.pickupAt();
    if (pickupAt !== null) {
      create.mutate({ channel, pickup_at: pickupAt, items, ...customer });
    }
  }

  function submit(): void {
    if (lines.length === 0) {
      setProblem("Add at least one product or combo.");
      return;
    }
    if (!choice.ready || pricing) {
      setProblem("Choose a pickup time that suits the items.");
      return;
    }
    setProblem(null);
    if (mode === "guest") {
      void guestForm.handleSubmit((guest) => takeOrder({ guest }))();
      return;
    }
    if (!customerQuery.data) {
      setProblem("Find the customer's account first.");
      return;
    }
    takeOrder({ customer_id: customerQuery.data.id });
  }

  return (
    <>
      <DashboardProfileSection id="staff-order-customer" title="Customer" variant="plain">
        <StaffOrderCustomer
          channel={channel}
          customerQuery={customerQuery}
          guestForm={guestForm}
          mode={mode}
          onChannelChange={setChannel}
          onGuestSubmit={submit}
          onLookup={setLookedUp}
          onModeChange={setMode}
        />
      </DashboardProfileSection>

      <DashboardProfileSection id="staff-order-items" title="Items" variant="plain">
        <StaffOrderItemsPicker disabled={create.isPending} lines={lines} onChange={setLines} />
        {quoteQuery.isError ? (
          <p className="text-sm text-error" role="status">
            {isApiError(quoteQuery.error) ? quoteQuery.error.message : "The items could not be priced."}
          </p>
        ) : null}
      </DashboardProfileSection>

      {priced ? (
        <DashboardProfileSection id="staff-order-pickup" title="Pickup" variant="plain">
          <div className="staff-new-order__pickup">
            <PickupSlotPicker {...choice.picker} errorId={PICKUP_ERROR_ID} />
            <p className="storefront-pickup-picker__error text-caption" id={PICKUP_ERROR_ID} role="status">
              {choice.message}
            </p>
            {choice.retry ? (
              <Button className="justify-self-start" type="button" variant="outline" onClick={choice.retry}>
                Try again
              </Button>
            ) : null}
          </div>
        </DashboardProfileSection>
      ) : null}

      <div className="staff-new-order__submit">
        <Button aria-busy={create.isPending || undefined} disabled={create.isPending} type="button" onClick={submit}>
          {create.isPending ? "Taking the order…" : "Take order"}
        </Button>
        {priced ? (
          <p aria-busy={pricing || undefined} className={pricing ? "text-sm text-muted" : "text-sm"}>
            Cash due at pickup{" "}
            <span className="text-tabular">{formatPriceCents(quoteQuery.data?.total_cents ?? 0)}</span>
          </p>
        ) : null}
        <p className="text-sm text-error" role="status">
          {problem}
        </p>
      </div>
    </>
  );
}
