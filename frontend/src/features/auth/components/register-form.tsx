"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { mapValidationDetailsToFormErrors } from "@/lib/validation";

import { useRegister } from "../hooks";
import { registerSchema, type RegisterInput } from "../schemas";
import { AuthInlineLink } from "./auth-inline-link";
import { AuthFormShell } from "./auth-form-shell";
import { AuthPasswordChecklist } from "./auth-password-checklist";

export function RegisterForm() {
  const registerMutation = useRegister();

  const form = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  const password = useWatch({
    control: form.control,
    name: "password",
    defaultValue: "",
  });

  function onSubmit(values: RegisterInput) {
    registerMutation.mutate(values, {
      onError: (error) => {
        if (!isApiError(error)) {
          toast.error("Something went wrong. Please try again.");
          return;
        }
        if (error.isEmailExists()) {
          form.setError("email", { message: "Email already registered" });
          return;
        }
        if (error.status === 422 && error.details) {
          for (const item of mapValidationDetailsToFormErrors(error.details)) {
            if (item.field === "email" || item.field === "password") {
              form.setError(item.field, { message: item.message });
            }
          }
          return;
        }
        toast.error(error.message);
      },
    });
  }

  return (
    <AuthFormShell
      description="Create your account to order for pickup."
      footer={
        <p className="auth-page-switch">
          Already have an account?{" "}
          <AuthInlineLink href={ROUTE.login}>Sign in</AuthInlineLink>
        </p>
      }
      title="Create account"
    >
      <Form {...form}>
        <form
          className="auth-form"
          onSubmit={form.handleSubmit(onSubmit)}
          noValidate
        >
          <FormField
            control={form.control}
            name="email"
            render={({ field }) => (
              <FormItem className="auth-field">
                <FieldControl label="Email">
                  <Input
                    autoComplete="email"
                    inputMode="email"
                    placeholder="you@example.com"
                    type="email"
                    {...field}
                  />
                </FieldControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="password"
            render={({ field }) => (
              <FormItem className="auth-field">
                <FieldControl label="Password">
                  <PasswordInput
                    autoComplete="new-password"
                    placeholder="••••••••"
                    {...field}
                  />
                </FieldControl>
                <AuthPasswordChecklist password={password} />
                <FormMessage />
              </FormItem>
            )}
          />

          <Button
            className="auth-submit w-full"
            disabled={registerMutation.isPending}
            type="submit"
          >
            {registerMutation.isPending
              ? "Creating account…"
              : "Create account"}
          </Button>
        </form>
      </Form>
    </AuthFormShell>
  );
}
