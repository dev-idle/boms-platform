"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

import { StorefrontErrorState } from "@/components/layouts/storefront-state";
import { BRAND } from "@/constants/brand";
import { USER_ROLE } from "@/constants/roles";
import { CheckboxField } from "@/components/ui/checkbox-field";
import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { PhoneInput } from "@/components/ui/phone-input";
import { useRemountingFormSnapshot } from "@/lib/hooks/use-remounting-form-snapshot";
import { applyApiFormFieldErrors } from "@/lib/validation";

import {
  customerSelfProfileFormSchema,
  type CustomerSelfProfileFormValues,
} from "../schemas/index";
import { useMe, useUpdateProfile } from "../hooks";
import {
  customerProfileSnapshot,
  normalizeCustomerProfileFormValues,
} from "../lib/profile-form-values";
import { AccountIdentityGroup } from "./account-identity-group";

type CustomerAccountProfileFormBodyProps = {
  email: string;
  initialValues: CustomerSelfProfileFormValues;
  onSaved: (values: CustomerSelfProfileFormValues) => void;
};

function CustomerAccountProfileFormBody({
  email,
  initialValues,
  onSaved,
}: CustomerAccountProfileFormBodyProps) {
  const updateProfile = useUpdateProfile();

  const form = useForm<CustomerSelfProfileFormValues>({
    resolver: zodResolver(customerSelfProfileFormSchema),
    defaultValues: initialValues,
  });

  function onSubmit(values: CustomerSelfProfileFormValues): void {
    const payload = normalizeCustomerProfileFormValues(values);

    updateProfile.mutate(
      {
        display_name: payload.display_name,
        phone: payload.phone,
        // Sent only when changed here: a box ticked when the page opened must not
        // undo an unsubscribe made since from an email.
        marketing_opt_in:
          payload.marketing_opt_in === initialValues.marketing_opt_in ? undefined : payload.marketing_opt_in,
      },
      {
        onError: (error) => {
          applyApiFormFieldErrors(form, error, ["display_name", "phone"], "We could not save your profile. Please try again.");
        },
        onSuccess: (me) => {
          if (me.role !== USER_ROLE.customer) {
            return;
          }
          onSaved(customerProfileSnapshot(me.profile));
        },
      },
    );
  }

  return (
    <Form {...form}>
      <form
        method="post"
        className="dashboard-profile-form storefront-account-form"
        noValidate
        onSubmit={form.handleSubmit(onSubmit)}
      >
        <AccountIdentityGroup email={email} variant="customer" />

        <FormField
          control={form.control}
          name="display_name"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Display name" optional>
                <Input
                  autoComplete="nickname"
                  placeholder="Your display name"
                  {...field}
                  value={field.value ?? ""}
                />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="phone"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Phone" optional>
                <PhoneInput {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="marketing_opt_in"
          render={({ field }) => (
            <FormItem>
              <CheckboxField
                checked={field.value}
                name={field.name}
                ref={field.ref}
                onBlur={field.onBlur}
                onChange={(event) => field.onChange(event.target.checked)}
              >
                Email me promotions from {BRAND.name}: new bakes and offers, now and then. You can stop them at any
                time.
              </CheckboxField>
            </FormItem>
          )}
        />

        <div className="dashboard-profile-form-actions">
          <DashboardFormSaveButton
            idleLabel="Save changes"
            isPending={updateProfile.isPending}
            pendingLabel="Saving…"
          />
        </div>
      </form>
    </Form>
  );
}

type CustomerAccountProfileFormProps = {
  email: string;
  initialSnapshot: CustomerSelfProfileFormValues;
};

function CustomerAccountProfileFormInner({
  email,
  initialSnapshot,
}: CustomerAccountProfileFormProps) {
  const { commitSnapshot, formKey, snapshot } =
    useRemountingFormSnapshot(initialSnapshot);

  return (
    <CustomerAccountProfileFormBody
      key={formKey}
      email={email}
      initialValues={snapshot}
      onSaved={commitSnapshot}
    />
  );
}

export function CustomerAccountProfileForm() {
  const me = useMe();

  if (me.isPending) {
    return (
      <div aria-busy="true" className="grid gap-5">
        <span className="skeleton block h-24 w-full" />
        <span className="skeleton block h-12 w-full" />
        <span className="skeleton block h-12 w-full" />
      </div>
    );
  }

  if (me.isError) {
    return <StorefrontErrorState onRetry={() => void me.refetch()} title="We could not load your profile" />;
  }

  // The customer gate admits only customers, so another role here is a bug.
  if (me.data.role !== USER_ROLE.customer) {
    throw new Error("The customer profile form rendered for another role");
  }

  return (
    <CustomerAccountProfileFormInner
      key={me.data.id}
      email={me.data.email}
      initialSnapshot={customerProfileSnapshot(me.data.profile)}
    />
  );
}
