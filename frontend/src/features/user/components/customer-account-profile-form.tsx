"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

import { USER_ROLE } from "@/constants/roles";
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
  customerSelfProfileFormSchema,
  type CustomerSelfProfileFormValues,
} from "../schemas/index";
import { useMe, useUpdateProfile } from "../hooks";
import type { CustomerProfile } from "../types";

import {
  applyCustomerSelfProfileFormErrors,
  customerProfileFormDefaults,
  nullableString,
} from "./helpers";

type CustomerAccountProfileFormBodyProps = {
  profile: CustomerProfile;
};

function CustomerAccountProfileFormBody({
  profile,
}: CustomerAccountProfileFormBodyProps) {
  const updateProfile = useUpdateProfile();

  const form = useForm<CustomerSelfProfileFormValues>({
    resolver: zodResolver(customerSelfProfileFormSchema),
    defaultValues: customerProfileFormDefaults(profile),
  });

  function onSubmit(values: CustomerSelfProfileFormValues): void {
    updateProfile.mutate(
      {
        display_name: values.display_name?.trim() || undefined,
        phone: nullableString(values.phone ?? ""),
      },
      {
        onError: (error) => {
          applyCustomerSelfProfileFormErrors(form, error, [
            "display_name",
            "phone",
          ]);
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

export function CustomerAccountProfileForm() {
  const me = useMe();

  if (me.isPending) {
    return <p className="text-sm text-muted">Loading profile…</p>;
  }

  if (!me.data || me.data.role !== USER_ROLE.customer) {
    return <p className="text-sm text-muted">Customer profile not available.</p>;
  }

  return (
    <CustomerAccountProfileFormBody
      key={me.data.id}
      profile={me.data.profile}
    />
  );
}
