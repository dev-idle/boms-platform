"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useForm, useWatch } from "react-hook-form";
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
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";

import { useRegister } from "../hooks";
import { mapValidationDetailsToFormErrors } from "../lib/validation-messages";
import { registerSchema, type RegisterInput } from "../schemas";
import { AuthenticatedRedirect } from "./authenticated-redirect";

function PasswordChecklist({ password }: { password: string }) {
  const checks = [
    { label: "At least 8 characters", ok: password.length >= 8 },
    { label: "Uppercase letter", ok: /[A-Z]/.test(password) },
    { label: "Digit", ok: /\d/.test(password) },
  ];

  return (
    <ul className="space-y-1 text-xs text-zinc-600 dark:text-zinc-400">
      {checks.map((check) => (
        <li
          key={check.label}
          className={
            check.ok ? "text-emerald-600 dark:text-emerald-400" : undefined
          }
        >
          {check.ok ? "✓" : "○"} {check.label}
        </li>
      ))}
    </ul>
  );
}

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
        if (error.code === "EMAIL_EXISTS") {
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
    <>
      <AuthenticatedRedirect />
      <div className="mx-auto flex min-h-full max-w-md flex-col justify-center px-6 py-16">
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
          Create account
        </h1>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
          Password rules mirror the backend validator.
        </p>

        <Form {...form}>
          <form
            className="mt-8 space-y-4"
            onSubmit={form.handleSubmit(onSubmit)}
            noValidate
          >
            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Email</FormLabel>
                  <FormControl>
                    <Input
                      autoComplete="email"
                      inputMode="email"
                      placeholder="you@example.com"
                      type="email"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Password</FormLabel>
                  <FormControl>
                    <Input
                      autoComplete="new-password"
                      placeholder="••••••••"
                      type="password"
                      {...field}
                    />
                  </FormControl>
                  <PasswordChecklist password={password} />
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button
              className="w-full"
              disabled={registerMutation.isPending}
              type="submit"
            >
              {registerMutation.isPending
                ? "Creating account…"
                : "Create account"}
            </Button>
          </form>
        </Form>

        <Link
          className="mt-8 text-sm font-medium text-zinc-900 underline dark:text-zinc-100"
          href={ROUTE.login}
        >
          Already registered?
        </Link>
      </div>
    </>
  );
}
