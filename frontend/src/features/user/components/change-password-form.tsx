"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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
import { newPasswordZodString } from "@/lib/validation/password";

import { useChangePassword } from "../hooks";

import { applyValidationDetails } from "./helpers";

const changePasswordFormSchema = z
  .object({
    old_password: z.string().min(1, "Current password is required"),
    new_password: newPasswordZodString(),
    confirm_password: z.string().min(1, "Confirm your new password"),
  })
  .superRefine((input, ctx) => {
    if (input.new_password !== input.confirm_password) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Passwords do not match",
        path: ["confirm_password"],
      });
    }
  });

type ChangePasswordFormValues = z.infer<typeof changePasswordFormSchema>;

export function ChangePasswordForm() {
  const mutation = useChangePassword();
  const form = useForm<ChangePasswordFormValues>({
    resolver: zodResolver(changePasswordFormSchema),
    defaultValues: {
      old_password: "",
      new_password: "",
      confirm_password: "",
    },
  });

  function onSubmit(values: ChangePasswordFormValues): void {
    mutation.mutate(
      {
        old_password: values.old_password,
        new_password: values.new_password,
      },
      {
        onError: (error) => {
          if (!isApiError(error)) {
            toast.error("Failed to change password");
            return;
          }
          if (error.status === 422 && error.details) {
            applyValidationDetails(error.details, (field, message) => {
              if (field in values) {
                form.setError(field as keyof ChangePasswordFormValues, {
                  message,
                });
              }
            });
            return;
          }
          if (error.code === "INVALID_CREDENTIALS") {
            form.setError("old_password", {
              message: "Current password is incorrect",
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
          name="old_password"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Current password</FormLabel>
              <FormControl>
                <Input autoComplete="current-password" type="password" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="new_password"
          render={({ field }) => (
            <FormItem>
              <FormLabel>New password</FormLabel>
              <FormControl>
                <Input autoComplete="new-password" type="password" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="confirm_password"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Confirm password</FormLabel>
              <FormControl>
                <Input autoComplete="new-password" type="password" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button disabled={mutation.isPending} type="submit">
          {mutation.isPending ? "Changing…" : "Change password"}
        </Button>
      </form>
    </Form>
  );
}
