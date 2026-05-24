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
import { validateNext } from "@/lib/validate-next";

import { useLogin } from "../hooks";
import { mapValidationDetailsToFormErrors } from "../lib/validation-messages";
import { loginSchema, type LoginInput } from "../schemas";
import { AuthenticatedRedirect } from "./authenticated-redirect";

type LoginFormProps = {
  next?: string;
  registered?: boolean;
};

export function LoginForm({ next, registered }: LoginFormProps) {
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

  function onSubmit(values: LoginInput) {
    login.mutate(
      { input: values, next },
      {
        onError: (error) => {
          if (!isApiError(error)) {
            toast.error("Something went wrong. Please try again.");
            return;
          }
          if (error.code === "INVALID_CREDENTIALS") {
            form.setError("root", {
              message: "Invalid credentials",
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
    <>
      <AuthenticatedRedirect
        to={validateNext(next ?? null) ?? ROUTE.home}
      />
      <div className="mx-auto flex min-h-full max-w-md flex-col justify-center px-6 py-16">
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
          Sign in
        </h1>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
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
              <p className="text-sm font-medium text-red-600 dark:text-red-400">
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
          className="mt-8 text-sm font-medium text-zinc-900 underline dark:text-zinc-100"
          href={ROUTE.register}
        >
          Create an account
        </Link>
      </div>
    </>
  );
}
