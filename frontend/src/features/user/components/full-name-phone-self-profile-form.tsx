"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { ReactNode } from "react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
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
              <FormLabel>Full name</FormLabel>
              <FormControl>
                <Input placeholder="Full name" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="phone"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Phone</FormLabel>
              <FormControl>
                <Input placeholder="Optional phone number" {...field} />
              </FormControl>
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
