"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";

import { useRequestPasswordReset } from "../hooks";
import { AUTH_FORM_COPY } from "../lib/auth-form-copy";
import { focusOnMount } from "../lib/focus-on-mount";
import { forgotPasswordSchema, type ForgotPasswordInput } from "../schemas";
import { AuthFormShell } from "./auth-form-shell";

const SIGN_IN_FOOTER = { href: ROUTE.login, linkLabel: "Sign in", prompt: "Remember your password?" };

/**
 * Asks for a reset link. The page answers the same whether or not the address
 * has an account, so it tells nobody who shops here.
 */
export function ForgotPasswordForm() {
  const request = useRequestPasswordReset();
  const form = useForm<ForgotPasswordInput>({
    resolver: zodResolver(forgotPasswordSchema),
    defaultValues: { email: "" },
  });

  function onSubmit(values: ForgotPasswordInput) {
    request.mutate(values, {
      onError: (error) => {
        toast.error(
          isApiError(error) && error.isRateLimited()
            ? "Too many requests from here. Try again in a few minutes."
            : "We could not send the link. Please try again.",
        );
      },
    });
  }

  if (request.isSuccess) {
    return (
      <AuthFormShell
        description={`If an account uses ${request.variables.email}, we sent it a link to choose a new password.`}
        footer={SIGN_IN_FOOTER}
        title="Check your inbox"
        titleRef={focusOnMount}
      >
        <Button className="w-full" type="button" variant="outline" onClick={() => request.reset()}>
          Use another address
        </Button>
      </AuthFormShell>
    );
  }

  return (
    <AuthFormShell
      description={AUTH_FORM_COPY.forgotPassword.description}
      footer={SIGN_IN_FOOTER}
      title={AUTH_FORM_COPY.forgotPassword.title}
    >
      <Form {...form}>
        <form method="post" className="auth-form" noValidate onSubmit={form.handleSubmit(onSubmit)}>
          <FormField
            control={form.control}
            name="email"
            render={({ field }) => (
              <FormItem className="auth-field">
                <FieldControl label="Email address">
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
          <Button
            aria-busy={request.isPending || undefined}
            className="auth-submit w-full"
            disabled={request.isPending}
            type="submit"
          >
            {request.isPending ? "Sending…" : "Send the link"}
          </Button>
        </form>
      </Form>
    </AuthFormShell>
  );
}
