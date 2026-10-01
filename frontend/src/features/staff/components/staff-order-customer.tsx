"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm, type UseFormReturn } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { PhoneInput } from "@/components/ui/phone-input";
import { ReadonlyFieldGroup } from "@/components/ui/readonly-field-group";
import { isApiError } from "@/lib/errors";
import { formatVietnamPhone } from "@/lib/validation/phone";

import type { useStaffCustomer } from "../hooks";
import { staffCustomerLookupSchema, type StaffCustomerLookup, type StaffGuestForm } from "../schemas";

export type StaffOrderChannel = "counter" | "phone";
export type StaffOrderCustomerMode = "guest" | "account";

const CHANNELS: Array<{ value: StaffOrderChannel; label: string }> = [
  { value: "counter", label: "At the counter" },
  { value: "phone", label: "On the phone" },
];

const CUSTOMER_MODES: Array<{ value: StaffOrderCustomerMode; label: string }> = [
  { value: "guest", label: "Guest" },
  { value: "account", label: "Customer account" },
];

type StaffOrderCustomerProps = {
  channel: StaffOrderChannel;
  onChannelChange: (channel: StaffOrderChannel) => void;
  mode: StaffOrderCustomerMode;
  onModeChange: (mode: StaffOrderCustomerMode) => void;
  customerQuery: ReturnType<typeof useStaffCustomer>;
  onLookup: (email: string) => void;
  guestForm: UseFormReturn<StaffGuestForm>;
  /** Enter in the guest's fields takes the order. */
  onGuestSubmit: () => void;
};

/** Where an order is taken and who it is for: a customer account found by its email, or a guest. */
export function StaffOrderCustomer({
  channel,
  onChannelChange,
  mode,
  onModeChange,
  customerQuery,
  onLookup,
  guestForm,
  onGuestSubmit,
}: StaffOrderCustomerProps) {
  const lookupForm = useForm<StaffCustomerLookup>({
    resolver: zodResolver(staffCustomerLookupSchema),
    defaultValues: { email: "" },
  });

  return (
    <div className="staff-new-order__customer">
      <div className="staff-new-order__choices">
        <div className="staff-new-order__choice">
          <span aria-hidden="true" className="text-form-label">
            Taken
          </span>
          <DashboardFilterGroup aria-label="Taken" onChange={onChannelChange} options={CHANNELS} value={channel} />
        </div>
        <div className="staff-new-order__choice">
          <span aria-hidden="true" className="text-form-label">
            For
          </span>
          <DashboardFilterGroup aria-label="For" onChange={onModeChange} options={CUSTOMER_MODES} value={mode} />
        </div>
      </div>

      {mode === "account" ? (
        <>
          <Form {...lookupForm}>
            <form
              className="staff-new-order__lookup"
              method="post"
              noValidate
              onSubmit={lookupForm.handleSubmit(({ email }) => onLookup(email))}
            >
              <FormField
                control={lookupForm.control}
                name="email"
                render={({ field }) => (
                  <FormItem className="contents">
                    <FieldControl label="Customer email">
                      <Input autoComplete="off" type="email" {...field} />
                    </FieldControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button disabled={customerQuery.isFetching} type="submit" variant="outline">
                {customerQuery.isFetching ? "Finding…" : "Find"}
              </Button>
            </form>
          </Form>
          <div aria-live="polite" role="status">
            {customerQuery.data ? (
              <ReadonlyFieldGroup
                heading="Customer account"
                reason="The customer changes these from their own account."
                rows={[
                  { label: "Name", value: customerQuery.data.display_name ?? "—" },
                  { label: "Email", value: customerQuery.data.email },
                  { label: "Phone", value: formatVietnamPhone(customerQuery.data.phone) || "—" },
                ]}
              />
            ) : customerQuery.isError ? (
              <p className="text-sm text-error">
                {isApiError(customerQuery.error) && customerQuery.error.status === 404
                  ? "No customer account has that email."
                  : "The account could not be looked up. Try again."}
              </p>
            ) : null}
          </div>
        </>
      ) : (
        <Form {...guestForm}>
          <form
            className="dashboard-profile-form"
            method="post"
            noValidate
            onSubmit={(event) => {
              event.preventDefault();
              onGuestSubmit();
            }}
          >
            <FormField
              control={guestForm.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FieldControl label="Name">
                    <Input autoComplete="off" maxLength={100} {...field} />
                  </FieldControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={guestForm.control}
              name="phone"
              render={({ field }) => (
                <FormItem>
                  <FieldControl label="Phone">
                    <PhoneInput autoComplete="off" {...field} />
                  </FieldControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </form>
        </Form>
      )}
    </div>
  );
}
