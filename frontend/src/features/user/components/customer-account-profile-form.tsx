"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { toast } from "sonner";

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
import { isApiError } from "@/lib/errors";

import { useMe, useUpdateProfile } from "../hooks";

import {
  applyValidationDetails,
  fieldValueOrEmpty,
  nullableString,
} from "./helpers";

const customerProfileFormSchema = z.object({
  display_name: z.string().trim().max(255).optional(),
  phone: z.string().trim().max(50).optional(),
});

type CustomerProfileFormValues = z.infer<typeof customerProfileFormSchema>;

export function CustomerAccountProfileForm() {
  const me = useMe();
  const updateProfile = useUpdateProfile();

  const form = useForm<CustomerProfileFormValues>({
    resolver: zodResolver(customerProfileFormSchema),
    defaultValues: { display_name: "", phone: "" },
  });

  useEffect(() => {
    if (me.data?.role !== "customer") {
      return;
    }
    form.reset({
      display_name: fieldValueOrEmpty(me.data.profile.display_name),
      phone: fieldValueOrEmpty(me.data.profile.phone),
    });
  }, [form, me.data]);

  if (me.isPending) {
    return <p className="text-sm text-zinc-500">Loading profile…</p>;
  }

  if (!me.data || me.data.role !== "customer") {
    return <p className="text-sm text-zinc-500">Customer profile not available.</p>;
  }

  function onSubmit(values: CustomerProfileFormValues): void {
    updateProfile.mutate(
      {
        display_name: values.display_name?.trim() || undefined,
        phone: nullableString(values.phone ?? ""),
      },
      {
        onError: (error) => {
          if (!isApiError(error)) {
            toast.error("Failed to update profile");
            return;
          }
          if (error.status === 422 && error.details) {
            applyValidationDetails(error.details, (field, message) => {
              if (field === "display_name" || field === "phone") {
                form.setError(field, { message });
              }
            });
            return;
          }
          toast.error(error.message);
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
          name="display_name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Display name</FormLabel>
              <FormControl>
                <Input placeholder="Your display name" {...field} />
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

        <Button disabled={updateProfile.isPending} type="submit">
          {updateProfile.isPending ? "Saving…" : "Save changes"}
        </Button>
      </form>
    </Form>
  );
}
