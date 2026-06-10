"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
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

import { useLogin } from "../hooks";
import { loginSchema, type LoginInput } from "../schemas";
import { AuthPageShell } from "./auth-page-shell";

type LoginFormProps = {
  next?: string;
  registered?: boolean;
  changed?: boolean;
};

export function LoginForm({ next, registered, changed }: LoginFormProps) {
  const login = useLogin();

  const form = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  useEffect(() => {
    if (registered) {
      toast.success("Account created. Sign in to continue.");
    }
  }, [registered]);

  useEffect(() => {
    if (changed) {
      toast.success("Password changed. Sign in again.");
    }
  }, [changed]);

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
    <AuthPageShell
      brandDescription="Sign in to browse our menu, schedule pickup, and follow your order from oven to counter."
      brandTitle="Freshly baked, ready when you are."
      description="Welcome back. Enter your details to continue."
      footer={
        <p className="text-sm text-muted">
          New here?{" "}
          <Link
            className="font-medium text-rose-500 underline-offset-4 transition-colors duration-standard ease-default hover:text-rose-600 hover:underline"
            href={ROUTE.register}
          >
            Create an account
          </Link>
        </p>
      }
      title="Sign in"
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
                    autoComplete="current-password"
                    placeholder="••••••••"
                    type="password"
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          {form.formState.errors.root?.message ? (
            <p className="text-sm font-medium text-error" role="alert">
              {form.formState.errors.root.message}
            </p>
          ) : null}

          <Button
            className="w-full"
            disabled={login.isPending}
            size="lg"
            type="submit"
          >
            {login.isPending ? "Signing in…" : "Sign in"}
          </Button>
        </form>
      </Form>
    </AuthPageShell>
  );
}
