"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
import { mapValidationDetailsToFormErrors } from "@/lib/validation";
import { registerHrefWithNext } from "@/lib/validate-next";

import { showLoginFlashToast } from "../lib/login-flash";
import { useLogin } from "../hooks";
import { loginSchema, type LoginInput } from "../schemas";
import { AuthInlineLink } from "./auth-inline-link";
import { AuthFormShell } from "./auth-form-shell";

type LoginFormProps = {
  next?: string;
  registered?: boolean;
  changed?: boolean;
};

export function LoginForm({ next, registered, changed }: LoginFormProps) {
  const login = useLogin();
  const router = useRouter();
  const searchParams = useSearchParams();

  const form = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  useEffect(() => {
    if (!registered && !changed) {
      return;
    }

    if (changed) {
      showLoginFlashToast("passwordChanged");
    } else if (registered) {
      showLoginFlashToast("accountRegistered");
    }

    const params = new URLSearchParams(searchParams.toString());
    params.delete("changed");
    params.delete("registered");
    const query = params.toString();
    router.replace(query ? `${ROUTE.login}?${query}` : ROUTE.login);
  }, [changed, registered, router, searchParams]);

  function onSubmit(values: LoginInput) {
    login.mutate(
      { input: values, next },
      {
        onError: (error) => {
          if (!isApiError(error)) {
            toast.error("Something went wrong. Please try again.");
            return;
          }
          if (error.isInvalidCredentials()) {
            form.setError("root", {
              message: "Invalid email or password",
            });
            return;
          }
          if (error.hasValidationDetails()) {
            for (const item of mapValidationDetailsToFormErrors(error.details!)) {
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
      description="Enter your email and password to continue."
      footer={
        <p className="auth-page-switch">
          New here?{" "}
          <AuthInlineLink href={registerHrefWithNext(next)}>
            Create an account
          </AuthInlineLink>
        </p>
      }
      title={PAGE_TITLES.signIn}
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
                    autoComplete="current-password"
                    placeholder="••••••••"
                    {...field}
                  />
                </FieldControl>
                <div className="auth-forgot-row">
                  <AuthInlineLink
                    className="auth-forgot-link"
                    href={ROUTE.forgotPassword}
                  >
                    Forgot password?
                  </AuthInlineLink>
                </div>
                <FormMessage />
              </FormItem>
            )}
          />

          {form.formState.errors.root?.message ? (
            <p className="text-caption text-error" role="alert">
              {form.formState.errors.root.message}
            </p>
          ) : null}

          <Button
            className="auth-submit w-full"
            disabled={login.isPending}
            type="submit"
          >
            {login.isPending ? "Signing in…" : "Sign in"}
          </Button>
        </form>
      </Form>
    </AuthFormShell>
  );
}
