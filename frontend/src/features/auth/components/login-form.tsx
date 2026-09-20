"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { mapValidationDetailsToFormErrors } from "@/lib/validation";
import { registerHrefWithNext, validateNext } from "@/lib/validate-next";

import { AUTH_FORM_COPY } from "../lib/auth-form-copy";
import { showLoginFlashToast } from "../lib/login-flash";
import { useLogin } from "../hooks";
import { loginSchema, type LoginInput } from "../schemas";
import { AuthFormShell } from "./auth-form-shell";

/** Reads `next`, `registered` and `changed` from the URL — render it inside <Suspense>. */
export function LoginForm() {
  const login = useLogin();
  const router = useRouter();
  const searchParams = useSearchParams();
  const next = validateNext(searchParams.get("next")) ?? undefined;
  const registered = searchParams.get("registered") === "1";
  const changed = searchParams.get("changed") === "1";
  // UI-only until the refresh-cookie lifetime becomes a login option on the API;
  // the value is deliberately not sent with the credentials.
  const [keepSignedIn, setKeepSignedIn] = useState(true);

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
      description={AUTH_FORM_COPY.signIn.description}
      footer={{
        href: registerHrefWithNext(next),
        linkLabel: "Create an account",
        prompt: "New to Choux?",
      }}
      footerNote={
        <>
          Collecting an order only?{" "}
          <Link className="auth-page-guest__link" href={ROUTE.products}>
            Continue as a guest
          </Link>
        </>
      }
      title={AUTH_FORM_COPY.signIn.title}
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
                <FieldControl
                  label="Password"
                  labelAction={
                    <Link className="auth-forgot-link" href={ROUTE.forgotPassword}>
                      Forgot?
                    </Link>
                  }
                >
                  <PasswordInput
                    autoComplete="current-password"
                    placeholder="••••••••"
                    {...field}
                  />
                </FieldControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <label className="auth-remember">
            <input
              checked={keepSignedIn}
              className="auth-remember__input"
              onChange={(event) => setKeepSignedIn(event.target.checked)}
              type="checkbox"
            />
            Keep me signed in on this device
          </label>

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
