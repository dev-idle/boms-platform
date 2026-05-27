"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { toast } from "sonner";

import { USER_ROLE, type UserRole } from "@/constants/roles";
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

const operationalSelfProfileFormSchema = z.object({
  full_name: z.string().trim().min(1, "Full name is required").max(255),
  phone: z.string().trim().max(50).optional(),
});

type OperationalSelfProfileFormValues = z.infer<
  typeof operationalSelfProfileFormSchema
>;

type ReadonlyStaffProfile = {
  employee_code: string;
  hire_date: string;
  shift: string;
};

const OPERATIONAL_ROLES = [
  USER_ROLE.staff,
  USER_ROLE.baker,
  USER_ROLE.manager,
] as const;

type OperationalRole = (typeof OPERATIONAL_ROLES)[number];

function isOperationalRole(role: UserRole): role is OperationalRole {
  return (OPERATIONAL_ROLES as readonly UserRole[]).includes(role);
}

function ReadonlyStaffFields({ profile }: { profile: ReadonlyStaffProfile }) {
  return (
    <div className="space-y-3 rounded-md border border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-800 dark:bg-zinc-950">
      <p className="text-xs font-medium uppercase tracking-wide text-zinc-500">
        Set by admin
      </p>
      <div className="grid gap-3 sm:grid-cols-3">
        <div>
          <p className="text-xs text-zinc-500">Employee code</p>
          <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
            {profile.employee_code}
          </p>
        </div>
        <div>
          <p className="text-xs text-zinc-500">Hire date</p>
          <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
            {profile.hire_date}
          </p>
        </div>
        <div>
          <p className="text-xs text-zinc-500">Shift</p>
          <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
            {profile.shift || "—"}
          </p>
        </div>
      </div>
    </div>
  );
}

type OperationalAccountProfileFormProps = {
  expectedRole: OperationalRole;
  roleLabel: string;
};

export function OperationalAccountProfileForm({
  expectedRole,
  roleLabel,
}: OperationalAccountProfileFormProps) {
  const me = useMe();
  const updateProfile = useUpdateProfile();

  const form = useForm<OperationalSelfProfileFormValues>({
    resolver: zodResolver(operationalSelfProfileFormSchema),
    defaultValues: { full_name: "", phone: "" },
  });

  useEffect(() => {
    if (!me.data || me.data.role !== expectedRole) {
      return;
    }
    form.reset({
      full_name: me.data.profile.full_name,
      phone: fieldValueOrEmpty(me.data.profile.phone),
    });
  }, [form, me.data, expectedRole]);

  if (me.isPending) {
    return <p className="text-sm text-zinc-500">Loading profile…</p>;
  }

  if (!me.data || me.data.role !== expectedRole) {
    return (
      <p className="text-sm text-zinc-500">{roleLabel} profile not available.</p>
    );
  }

  function onSubmit(values: OperationalSelfProfileFormValues): void {
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

  const readonlyProfile = isOperationalRole(me.data.role) ? me.data.profile : null;

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

        {readonlyProfile ? <ReadonlyStaffFields profile={readonlyProfile} /> : null}

        <Button disabled={updateProfile.isPending} type="submit">
          {updateProfile.isPending ? "Saving…" : "Save changes"}
        </Button>
      </form>
    </Form>
  );
}

/** @deprecated Use OperationalAccountProfileForm with expectedRole={USER_ROLE.staff}. */
export function StaffAccountProfileForm() {
  return (
    <OperationalAccountProfileForm
      expectedRole={USER_ROLE.staff}
      roleLabel="Staff"
    />
  );
}
