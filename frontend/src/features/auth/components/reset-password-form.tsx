"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { PasswordInput } from "@/components/ui/password-input";
import { ROUTE } from "@/constants/routes";
import { ApiErrorCode, isApiError } from "@/lib/errors";
import { applyFormFieldErrors } from "@/lib/validation";

import { useConfirmPasswordReset, useLinkToken } from "../hooks";
import { AUTH_FORM_COPY } from "../lib/auth-form-copy";
import { focusOnMount } from "../lib/focus-on-mount";
import { resetPasswordFormSchema, type ResetPasswordFormInput } from "../schemas";
import { AuthFormShell } from "./auth-form-shell";
import { AuthFormSkeleton } from "./auth-form-skeleton";
import { AuthPasswordChecklist } from "./auth-password-checklist";

const NEW_LINK_FOOTER = { href: ROUTE.forgotPassword, linkLabel: "Get a new link", prompt: "Link not working?" };

/** Sets a new password from an emailed reset link. */
export function ResetPasswordForm() {
  const token = useLinkToken();
  const confirm = useConfirmPasswordReset();
  const form = useForm<ResetPasswordFormInput>({
    resolver: zodResolver(resetPasswordFormSchema),
    defaultValues: { new_password: "", confirm_password: "" },
  });
  const password = useWatch({ control: form.control, name: "new_password" });

  if (token === undefined) {
    return <AuthFormSkeleton {...AUTH_FORM_COPY.resetPassword} />;
  }

  if (token === null || (isApiError(confirm.error) && confirm.error.code === ApiErrorCode.InvalidLink)) {
    return (
      <AuthFormShell
        description="This link has expired or has already been used."
        title="This link no longer works"
        titleRef={focusOnMount}
      >
        <Button asChild className="w-full">
          <Link href={ROUTE.forgotPassword}>Get a new link</Link>
        </Button>
      </AuthFormShell>
    );
  }

  function onSubmit(values: ResetPasswordFormInput) {
    if (!token) {
      return;
    }
    confirm.mutate(
      { token, new_password: values.new_password },
      {
        onError: (error) => {
          if (isApiError(error) && error.hasValidationDetails()) {
            applyFormFieldErrors(form, error.details!, ["new_password"]);
            return;
          }
          if (isApiError(error) && error.code === ApiErrorCode.InvalidLink) {
            return;
          }
          toast.error(
            isApiError(error) && error.isRateLimited()
              ? "Too many attempts from here. Try again in a few minutes."
              : "We could not save your new password. Please try again.",
          );
        },
      },
    );
  }

  return (
    <AuthFormShell
      description={AUTH_FORM_COPY.resetPassword.description}
      footer={NEW_LINK_FOOTER}
      title={AUTH_FORM_COPY.resetPassword.title}
    >
      <Form {...form}>
        <form className="auth-form" noValidate onSubmit={form.handleSubmit(onSubmit)}>
          <FormField
            control={form.control}
            name="new_password"
            render={({ field }) => (
              <FormItem className="auth-field">
                <FieldControl label="New password">
                  <PasswordInput autoComplete="new-password" {...field} />
                </FieldControl>
                <AuthPasswordChecklist password={password} />
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="confirm_password"
            render={({ field }) => (
              <FormItem className="auth-field">
                <FieldControl label="Confirm password">
                  <PasswordInput autoComplete="new-password" {...field} />
                </FieldControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <Button
            aria-busy={confirm.isPending || undefined}
            className="auth-submit w-full"
            disabled={confirm.isPending}
            type="submit"
          >
            {confirm.isPending ? "Saving…" : "Save the new password"}
          </Button>
        </form>
      </Form>
    </AuthFormShell>
  );
}
