"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useSearchParams } from "next/navigation";
import { useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { CheckboxField } from "@/components/ui/checkbox-field";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { BRAND } from "@/constants/brand";
import { PolicyConsent } from "@/features/legal";
import { loginHrefWithNext, validateNext } from "@/lib/validate-next";
import { ApiErrorCode, isApiError } from "@/lib/errors";
import { mapValidationDetailsToFormErrors } from "@/lib/validation";

import { useRegister } from "../hooks";
import { AUTH_FORM_COPY } from "../lib/auth-form-copy";
import { registerSchema, type RegisterInput } from "../schemas";
import { AuthFormShell } from "./auth-form-shell";
import { AuthPasswordChecklist } from "./auth-password-checklist";

/** Reads `next` from the URL — render it inside <Suspense>. */
export function RegisterForm() {
  const registerMutation = useRegister();
  const next = validateNext(useSearchParams().get("next")) ?? undefined;

  const form = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: "",
      password: "",
      accept_terms: false,
      marketing_opt_in: false,
    },
  });

  const password = useWatch({
    control: form.control,
    name: "password",
    defaultValue: "",
  });

  function onSubmit(values: RegisterInput) {
    registerMutation.mutate(
      { input: values, next },
      {
        onError: (error) => {
          if (!isApiError(error)) {
            toast.error("Something went wrong. Please try again.");
            return;
          }
          if (error.isEmailExists()) {
            form.setError("email", { message: "Email already registered" });
            return;
          }
          // The policies changed while the form was open: the ones on screen are
          // no longer the ones in force.
          if (error.code === ApiErrorCode.TermsNotAccepted) {
            form.setValue("accept_terms", false);
            form.setError("accept_terms", {
              message: "Our policies were just updated. Reload the page to read and accept the current version.",
            });
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
      },
    );
  }

  return (
    <AuthFormShell
      description={AUTH_FORM_COPY.createAccount.description}
      footer={{
        href: loginHrefWithNext(next),
        linkLabel: "Sign in",
        prompt: "Already have an account?",
      }}
      title={AUTH_FORM_COPY.createAccount.title}
    >
      <Form {...form}>
        <form
          method="post"
          className="auth-form"
          onSubmit={form.handleSubmit(onSubmit)}
          noValidate
        >
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

          <FormField
            control={form.control}
            name="accept_terms"
            render={({ field, fieldState }) => (
              <FormItem className="auth-field">
                <PolicyConsent
                  checked={field.value}
                  describedBy={fieldState.error ? "register-terms-error" : undefined}
                  invalid={Boolean(fieldState.error)}
                  onCheckedChange={field.onChange}
                  scope="account"
                />
                <FormMessage id="register-terms-error" />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="marketing_opt_in"
            render={({ field }) => (
              <FormItem className="auth-field">
                <CheckboxField
                  checked={field.value}
                  name={field.name}
                  ref={field.ref}
                  onBlur={field.onBlur}
                  onChange={(event) => field.onChange(event.target.checked)}
                >
                  Email me promotions from {BRAND.name}: new bakes and offers, now and then. You can stop them at any
                  time.
                </CheckboxField>
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
