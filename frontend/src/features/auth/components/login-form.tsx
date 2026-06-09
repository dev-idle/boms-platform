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
import { useLogin } from "../hooks";
import { mapValidationDetailsToFormErrors } from "@/lib/validation";
import { loginSchema, type LoginInput } from "../schemas";

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
              message: "Invalid credentials",
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
    <div className="mx-auto flex min-h-full max-w-md flex-col justify-center px-6 py-16">
        <h1 className="font-heading text-2xl font-medium text-foreground">
          Sign in
        </h1>
        <p className="mt-2 text-sm text-muted">
          Access token stays in memory. Refresh is handled via HttpOnly cookie.
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
              <p className="text-sm font-medium text-error">
                {form.formState.errors.root.message}
              </p>
            ) : null}

            <Button
              className="w-full"
              disabled={login.isPending}
              type="submit"
            >
              {login.isPending ? "Signing in…" : "Sign in"}
            </Button>
          </form>
        </Form>

        <Link
          className="mt-8 text-sm font-medium text-foreground underline underline-offset-4"
          href={ROUTE.register}
        >
          Create an account
        </Link>
      </div>
  );
}
