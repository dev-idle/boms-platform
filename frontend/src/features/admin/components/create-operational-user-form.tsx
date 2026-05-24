"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm, useWatch } from "react-hook-form";
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
import { mapValidationDetailsToFormErrors } from "@/lib/validation";

import { useCreateOperational } from "../hooks";
import type { CreateOperationalInput } from "../schemas";

import { TempPasswordModal } from "./temp-password-modal";

const createOperationalUserFormSchema = z
  .object({
    email: z.string().trim().email("Enter a valid email").max(255),
    role: z.enum(["staff", "baker", "manager", "admin"]),
    full_name: z.string().trim().min(1, "Full name is required").max(255),
    phone: z.string().trim().max(50).optional(),
    employee_code: z.string().trim().max(64).optional(),
    hire_date: z.string().trim().optional(),
    shift: z.string().trim().max(64).optional(),
  })
  .superRefine((input, ctx) => {
    const isStaffRole =
      input.role === "staff" || input.role === "baker" || input.role === "manager";
    if (!isStaffRole) {
      return;
    }
    if (!input.employee_code) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Employee code is required",
        path: ["employee_code"],
      });
    }
    if (!input.hire_date) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Hire date is required",
        path: ["hire_date"],
      });
    }
    if (!input.shift) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Shift is required",
        path: ["shift"],
      });
    }
  });

type CreateOperationalUserFormValues = z.infer<
  typeof createOperationalUserFormSchema
>;

export function CreateOperationalUserForm() {
  const createUser = useCreateOperational();

  const form = useForm<CreateOperationalUserFormValues>({
    resolver: zodResolver(createOperationalUserFormSchema),
    defaultValues: {
      email: "",
      role: "staff",
      full_name: "",
      phone: "",
      employee_code: "",
      hire_date: "",
      shift: "",
    },
  });

  const role = useWatch({ control: form.control, name: "role" });
  const isStaffRole = role === "staff" || role === "baker" || role === "manager";

  function onSubmit(values: CreateOperationalUserFormValues): void {
    const payload: CreateOperationalInput =
      values.role === "admin"
        ? {
            email: values.email.trim(),
            role: "admin",
            full_name: values.full_name.trim(),
            phone: values.phone?.trim() ? values.phone.trim() : null,
          }
        : {
            email: values.email.trim(),
            role: values.role,
            full_name: values.full_name.trim(),
            phone: values.phone?.trim() ? values.phone.trim() : null,
            employee_code: values.employee_code?.trim() ?? "",
            hire_date: values.hire_date?.trim() ?? "",
            shift: values.shift?.trim() ?? "",
          };

    createUser.mutate(payload, {
      onError: (error) => {
        if (!isApiError(error)) {
          toast.error("Failed to create user");
          return;
        }
        if (error.status === 422 && error.details) {
          for (const item of mapValidationDetailsToFormErrors(error.details)) {
            if (item.field in values) {
              form.setError(
                item.field as keyof CreateOperationalUserFormValues,
                { message: item.message },
              );
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
          <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
            New operational user
          </h1>
          <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">
            Create staff, baker, manager, or admin users.
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
                      className="h-10 w-full rounded-md border border-zinc-200 bg-white px-3 text-sm dark:border-zinc-800 dark:bg-zinc-950"
                      {...field}
                    >
                      <option value="staff">staff</option>
                      <option value="baker">baker</option>
                      <option value="manager">manager</option>
                      <option value="admin">admin</option>
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
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {isStaffRole ? (
              <>
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

                <FormField
                  control={form.control}
                  name="hire_date"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Hire date</FormLabel>
                      <FormControl>
                        <Input type="date" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="shift"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Shift</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </>
            ) : null}

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
