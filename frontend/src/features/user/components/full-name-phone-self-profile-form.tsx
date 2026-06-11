"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { ReactNode } from "react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";

import {
  fullNamePhoneSelfProfileFormSchema,
  type FullNamePhoneSelfProfileFormValues,
} from "../schemas/index";
import { useUpdateProfile } from "../hooks";

import {
  applySelfProfileFormErrors,
  fullNamePhoneFormDefaults,
  nullableString,
} from "./helpers";

type FullNamePhoneSelfProfileFormProps = {
  fullName: string;
  phone: string | null | undefined;
  children?: ReactNode;
};

/** PATCH /me self-service form for roles that edit `full_name` and `phone`. */
export function FullNamePhoneSelfProfileForm({
  fullName,
  phone,
  children,
}: FullNamePhoneSelfProfileFormProps) {
  const updateProfile = useUpdateProfile();

  const form = useForm<FullNamePhoneSelfProfileFormValues>({
    resolver: zodResolver(fullNamePhoneSelfProfileFormSchema),
    defaultValues: fullNamePhoneFormDefaults(fullName, phone),
  });

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
        className="space-y-4"
        noValidate
        onSubmit={form.handleSubmit(onSubmit)}
      >
        <FormField
          control={form.control}
          name="full_name"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Full name">
                <Input placeholder="Full name" {...field} />
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
              <FieldControl label="Phone">
                <Input placeholder="Optional phone number" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {children}

        <Button disabled={updateProfile.isPending} type="submit">
          {updateProfile.isPending ? "Saving…" : "Save changes"}
        </Button>
      </form>
    </Form>
  );
}
