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

const adminProfileFormSchema = z.object({
  full_name: z.string().trim().min(1, "Full name is required").max(255),
  phone: z.string().trim().max(50).optional(),
});

type AdminProfileFormValues = z.infer<typeof adminProfileFormSchema>;

export function AdminAccountProfileForm() {
  const me = useMe();
  const updateProfile = useUpdateProfile();

  const form = useForm<AdminProfileFormValues>({
    resolver: zodResolver(adminProfileFormSchema),
    defaultValues: { full_name: "", phone: "" },
  });

  useEffect(() => {
    if (me.data?.role !== "admin") {
      return;
    }
    form.reset({
      full_name: me.data.profile.full_name,
      phone: fieldValueOrEmpty(me.data.profile.phone),
    });
  }, [form, me.data]);

  if (me.isPending) {
    return <p className="text-sm text-zinc-500">Loading profile…</p>;
  }

  if (!me.data || me.data.role !== "admin") {
    return <p className="text-sm text-zinc-500">Admin profile not available.</p>;
  }

  function onSubmit(values: AdminProfileFormValues): void {
    updateProfile.mutate(
      {
        full_name: values.full_name.trim(),
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
              if (field === "full_name" || field === "phone") {
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

        <Button disabled={updateProfile.isPending} type="submit">
          {updateProfile.isPending ? "Saving…" : "Save changes"}
        </Button>
      </form>
    </Form>
  );
}
