"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
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
import { ASSIGNABLE_OPERATIONAL_ROLES } from "@/constants/roles";
import { isApiError } from "@/lib/errors";
import { mapValidationDetailsToFormErrors } from "@/lib/validation";

import { useCreateOperational } from "../hooks";
import { createOperationalSchema, type CreateOperationalInput } from "../schemas";

import { TempPasswordModal } from "./temp-password-modal";

export function CreateOperationalUserForm() {
  const createUser = useCreateOperational();

  const form = useForm<CreateOperationalInput>({
    resolver: zodResolver(createOperationalSchema),
    defaultValues: {
      email: "",
      role: "staff",
      full_name: "",
      phone: null,
      employee_code: "",
    },
  });

  function onSubmit(values: CreateOperationalInput): void {
    createUser.mutate(values, {
      onError: (error) => {
        if (!isApiError(error)) {
          toast.error("Failed to create user");
          return;
        }
        if (error.isEmployeeCodeExists()) {
          toast.error("That employee code is already in use.");
          return;
        }
        if (error.hasValidationDetails()) {
          for (const item of mapValidationDetailsToFormErrors(error.details!)) {
            const field = item.field as keyof CreateOperationalInput;
            if (field in values) {
              form.setError(field, { message: item.message });
            }
          }
          return;
        }
        toast.error(error.message);
      },
    });
  }

  return (
    <>
      <div className="max-w-2xl space-y-6">
        <div>
          <h1 className="text-2xl font-medium text-ink">
            New operational user
          </h1>
          <p className="mt-2 text-sm text-ink-2">
            Create staff, baker, or manager accounts. Platform admins are created via dev seed only.
          </p>
        </div>

        <Form {...form}>
          <form
            className="space-y-4"
            noValidate
            onSubmit={form.handleSubmit(onSubmit)}
          >
            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Email</FormLabel>
                  <FormControl>
                    <Input type="email" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="role"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Role</FormLabel>
                  <FormControl>
                    <select
                      className="h-10 w-full rounded-md border border-border bg-surface px-3 text-sm"
                      {...field}
                    >
                      {ASSIGNABLE_OPERATIONAL_ROLES.map((role) => (
                        <option key={role} value={role}>
                          {role}
                        </option>
                      ))}
                    </select>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="full_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Full name</FormLabel>
                  <FormControl>
                    <Input {...field} />
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
                    <Input
                      {...field}
                      value={field.value ?? ""}
                      onChange={(e) => field.onChange(e.target.value || null)}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="employee_code"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Employee code</FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button disabled={createUser.isPending} type="submit">
              {createUser.isPending ? "Creating…" : "Create user"}
            </Button>
          </form>
        </Form>
      </div>

      <TempPasswordModal
        data={createUser.tempPasswordData}
        onClose={createUser.clearTempPasswordData}
        open={createUser.tempPasswordData !== null}
      />
    </>
  );
}
