"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { PasswordInput } from "@/components/ui/password-input";
import { isApiError } from "@/lib/errors";

import {
  changePasswordFormSchema,
  type ChangePasswordFormInput,
} from "../schemas/index";
import { useChangePassword } from "../hooks";

import { cn } from "@/lib/utils";

import { applyValidationDetails } from "./helpers";

type ChangePasswordFormProps = {
  formClassName?: string;
};

export function ChangePasswordForm({ formClassName }: ChangePasswordFormProps) {
  const mutation = useChangePassword();
  const form = useForm<ChangePasswordFormInput>({
    resolver: zodResolver(changePasswordFormSchema),
    defaultValues: {
      old_password: "",
      new_password: "",
      confirm_password: "",
    },
  });

  function onSubmit(values: ChangePasswordFormInput): void {
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
                form.setError(field as keyof ChangePasswordFormInput, {
                  message,
                });
              }
            });
            return;
          }
          if (error.isInvalidCredentials()) {
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
        className={cn("dashboard-profile-form", formClassName)}
        noValidate
        onSubmit={form.handleSubmit(onSubmit)}
      >
        <FormField
          control={form.control}
          name="old_password"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Current password">
                <PasswordInput autoComplete="current-password" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="new_password"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="New password">
                <PasswordInput autoComplete="new-password" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="confirm_password"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Confirm password">
                <PasswordInput autoComplete="new-password" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="dashboard-profile-form-actions">
          <Button disabled={mutation.isPending} type="submit">
            {mutation.isPending ? "Changing…" : "Change password"}
          </Button>
        </div>
      </form>
    </Form>
  );
}
