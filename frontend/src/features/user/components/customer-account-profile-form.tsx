"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

import { USER_ROLE } from "@/constants/roles";
import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
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
import { ReadonlyAccountEmailField } from "./readonly-account-email-field";

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
      },
      {
        onError: (error) => {
          applyApiFormFieldErrors(form, error, ["display_name", "phone"], "Failed to update profile");
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
        className="dashboard-profile-form storefront-account-form"
        noValidate
        onSubmit={form.handleSubmit(onSubmit)}
      >
        <ReadonlyAccountEmailField email={email} />

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
                <Input
                  autoComplete="tel"
                  inputMode="tel"
                  placeholder="Phone number"
                  type="tel"
                  {...field}
                  value={field.value ?? ""}
                />
              </FieldControl>
              <FormMessage />
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
    return <InlineLoadingState className="storefront-account-form__loading" />;
  }

  if (!me.data || me.data.role !== USER_ROLE.customer) {
    return <p className="text-sm text-muted">Customer profile not available.</p>;
  }

  return (
    <CustomerAccountProfileFormInner
      key={me.data.id}
      email={me.data.email}
      initialSnapshot={customerProfileSnapshot(me.data.profile)}
    />
  );
}
