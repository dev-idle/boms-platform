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
import { mapValidationDetailsToFormErrors } from "@/lib/validation";

import { useRegister } from "../hooks";
import { registerSchema, type RegisterInput } from "../schemas";
import { AuthPageShell } from "./auth-page-shell";

function PasswordChecklist({ password }: { password: string }) {
  const checks = [
    { label: "At least 8 characters", ok: password.length >= 8 },
    { label: "Uppercase letter", ok: /[A-Z]/.test(password) },
    { label: "Digit", ok: /\d/.test(password) },
  ];

  return (
    <ul aria-label="Password requirements" className="mt-2 space-y-1 text-xs">
      {checks.map((check) => (
        <li
          key={check.label}
          className={
            check.ok
              ? "text-success"
              : "text-muted"
          }
        >
          <span aria-hidden="true">{check.ok ? "✓" : "○"}</span> {check.label}
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
    <AuthPageShell
      brandDescription="Create an account to order for pickup, choose your time slot, and track every step until your treats are ready."
      brandTitle="Your next celebration starts here."
      description="A few details and you can start ordering for pickup."
      footer={
        <p className="text-sm text-muted">
          Already have an account?{" "}
          <Link
            className="font-medium text-rose-500 underline-offset-4 transition-colors duration-standard ease-default hover:text-rose-600 hover:underline"
            href={ROUTE.login}
          >
            Sign in
          </Link>
        </p>
      }
      title="Create account"
    >
      <Form {...form}>
        <form
          className="space-y-5"
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
            size="lg"
            type="submit"
          >
            {registerMutation.isPending
              ? "Creating account…"
              : "Create account"}
          </Button>
        </form>
      </Form>
    </AuthPageShell>
  );
}
