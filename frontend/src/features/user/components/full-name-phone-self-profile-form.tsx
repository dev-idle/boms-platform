"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { ReactNode } from "react";
import { useForm } from "react-hook-form";

import { USER_ROLE } from "@/constants/roles";
import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { useRemountingFormSnapshot } from "@/lib/hooks/use-remounting-form-snapshot";
import { applyApiFormFieldErrors } from "@/lib/validation";

import {
  fullNamePhoneSelfProfileFormSchema,
  type FullNamePhoneSelfProfileFormValues,
} from "../schemas/index";
import { useUpdateProfile } from "../hooks";
import {
  fullNamePhoneFormValuesEqual,
  fullNamePhoneSnapshotFromMe,
  normalizeFullNamePhoneFormValues,
} from "../lib/profile-form-values";

import { cn } from "@/lib/utils";

type FullNamePhoneSelfProfileFormProps = {
  initialSnapshot: FullNamePhoneSelfProfileFormValues;
  children?: ReactNode;
  formClassName?: string;
};

type FullNamePhoneSelfProfileFormBodyProps = {
  children?: ReactNode;
  formClassName?: string;
  initialValues: FullNamePhoneSelfProfileFormValues;
  onSaved: (values: FullNamePhoneSelfProfileFormValues) => void;
};

function FullNamePhoneSelfProfileFormBody({
  children,
  formClassName,
  initialValues,
  onSaved,
}: FullNamePhoneSelfProfileFormBodyProps) {
  const updateProfile = useUpdateProfile();

  const form = useForm<FullNamePhoneSelfProfileFormValues>({
    resolver: zodResolver(fullNamePhoneSelfProfileFormSchema),
    defaultValues: initialValues,
  });

  function onSubmit(values: FullNamePhoneSelfProfileFormValues): void {
    const payload = normalizeFullNamePhoneFormValues(values);

    updateProfile.mutate(
      {
        full_name: payload.full_name,
        phone: payload.phone,
      },
      {
        onError: (error) => {
          applyApiFormFieldErrors(form, error, ["full_name", "phone"], "Failed to update profile");
        },
        onSuccess: (me) => {
          if (me.role === USER_ROLE.customer) {
            return;
          }
          onSaved(fullNamePhoneSnapshotFromMe(me));
        },
      },
    );
  }

  return (
    <Form {...form}>
      <form
        className={cn("dashboard-profile-form", formClassName)}
        noValidate
        onSubmit={form.handleSubmit(onSubmit)}
      >
        <FormField
          control={form.control}
          name="full_name"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Full name">
                <Input
                  autoComplete="name"
                  placeholder="Full name"
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

        {children ? (
          <div className="dashboard-profile-form-span">{children}</div>
        ) : null}

        <div className="dashboard-profile-form-actions dashboard-profile-form-span">
          <DashboardFormSaveButton
            areEqual={fullNamePhoneFormValuesEqual}
            baseline={initialValues}
            form={form}
            idleLabel="Save changes"
            isPending={updateProfile.isPending}
            pendingLabel="Saving…"
          />
        </div>
      </form>
    </Form>
  );
}

/** PATCH /me self-service form for roles that edit `full_name` and `phone`. */
export function FullNamePhoneSelfProfileForm({
  initialSnapshot,
  children,
  formClassName,
}: FullNamePhoneSelfProfileFormProps) {
  const { commitSnapshot, formKey, snapshot } =
    useRemountingFormSnapshot(initialSnapshot);

  return (
    <FullNamePhoneSelfProfileFormBody
      key={formKey}
      formClassName={formClassName}
      initialValues={snapshot}
      onSaved={commitSnapshot}
    >
      {children}
    </FullNamePhoneSelfProfileFormBody>
  );
}
