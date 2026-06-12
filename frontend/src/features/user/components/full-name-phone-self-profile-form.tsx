"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { ReactNode } from "react";
import { useEffect, useRef } from "react";
import { useForm, useFormState } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";

import {
  fullNamePhoneSelfProfileFormSchema,
  type FullNamePhoneSelfProfileFormValues,
} from "../schemas/index";
import { useUpdateProfile } from "../hooks";

import { cn } from "@/lib/utils";

import {
  applySelfProfileFormErrors,
  fullNamePhoneFormDefaults,
  nullableString,
} from "./helpers";

type FullNamePhoneSelfProfileFormProps = {
  fullName: string;
  phone: string | null | undefined;
  children?: ReactNode;
  formClassName?: string;
  splitFields?: boolean;
};

/** PATCH /me self-service form for roles that edit `full_name` and `phone`. */
export function FullNamePhoneSelfProfileForm({
  fullName,
  phone,
  children,
  formClassName,
  splitFields = false,
}: FullNamePhoneSelfProfileFormProps) {
  const updateProfile = useUpdateProfile();

  const form = useForm<FullNamePhoneSelfProfileFormValues>({
    resolver: zodResolver(fullNamePhoneSelfProfileFormSchema),
    defaultValues: fullNamePhoneFormDefaults(fullName, phone),
  });

  const { reset, control } = form;
  const { isDirty } = useFormState({ control });
  const syncedServerValuesRef = useRef({ fullName, phone });

  // Sync when /me changes after save — never put the full `form` object in deps (wipes keystrokes).
  useEffect(() => {
    const prev = syncedServerValuesRef.current;
    const serverPhone = phone ?? null;
    const prevPhone = prev.phone ?? null;
    if (prev.fullName === fullName && prevPhone === serverPhone) {
      return;
    }
    syncedServerValuesRef.current = { fullName, phone };
    reset(fullNamePhoneFormDefaults(fullName, phone));
  }, [fullName, phone, reset]);

  function onSubmit(values: FullNamePhoneSelfProfileFormValues): void {
    updateProfile.mutate(
      {
        full_name: values.full_name.trim(),
        phone: nullableString(values.phone ?? ""),
      },
      {
        onError: (error) => {
          applySelfProfileFormErrors(form, error, ["full_name", "phone"]);
        },
      },
    );
  }

  return (
    <Form {...form}>
      <form
        className={cn(
          "dashboard-profile-form",
          splitFields && "dashboard-profile-form--split",
          formClassName,
        )}
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
                  placeholder="Phone number"
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
          <Button
            disabled={updateProfile.isPending || !isDirty}
            type="submit"
          >
            {updateProfile.isPending ? "Saving…" : "Save changes"}
          </Button>
        </div>
      </form>
    </Form>
  );
}
