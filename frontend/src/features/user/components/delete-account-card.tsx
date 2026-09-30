"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { PasswordInput } from "@/components/ui/password-input";
import { ApiErrorCode, isApiError } from "@/lib/errors";
import { applyFormFieldErrors } from "@/lib/validation";

import { useDeleteAccount, useMe } from "../hooks";
import { eraseMyAccountSchema, type EraseMyAccountInput } from "../schemas/index";

export function DeleteAccountCard() {
  const mutation = useDeleteAccount();
  const me = useMe();
  const [open, setOpen] = useState(false);
  const form = useForm<EraseMyAccountInput>({
    resolver: zodResolver(eraseMyAccountSchema),
    defaultValues: { password: "" },
  });

  const canDelete = me.data?.role === "customer";
  // A greyed action says why it is greyed (02-COMPONENTS §4): still loading, or
  // not a customer account.
  const disabledReason = me.isPending
    ? "Loading your account…"
    : canDelete
      ? undefined
      : "Only customer accounts can be deleted here.";

  function erase(): void {
    mutation.mutate(form.getValues(), {
      // A refusal closes the dialog and says why; the account is untouched.
      onError: (error) => {
        setOpen(false);
        if (isApiError(error) && error.hasValidationDetails()) {
          applyFormFieldErrors(form, error.details!, ["password"]);
          return;
        }
        if (isApiError(error) && error.isInvalidCredentials()) {
          form.setError("password", { message: "Current password is incorrect" });
          return;
        }
        toast.error(
          isApiError(error) && error.code === ApiErrorCode.AccountHasOpenOrders
            ? "You have an order that is not collected or cancelled yet. You can delete your account once it is."
            : "We could not delete your account. Please try again.",
        );
      },
    });
  }

  return (
    <>
      <Form {...form}>
        <form
          className="storefront-account-danger"
          noValidate
          onSubmit={form.handleSubmit(() => setOpen(true))}
        >
          <p className="storefront-account-danger__copy">
            We erase your name, phone number and email and sign you out everywhere.
            Your past orders stay in our sales records without your details. This
            cannot be undone, and waits until no order of yours is still open.
          </p>
          <FormField
            control={form.control}
            name="password"
            render={({ field }) => (
              <FormItem>
                <FieldControl
                  hint="Erasing cannot be undone, so we check it is you."
                  label="Current password"
                >
                  <PasswordInput autoComplete="current-password" {...field} />
                </FieldControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <Button
            aria-busy={me.isPending || undefined}
            disabled={!canDelete || mutation.isPending}
            title={disabledReason}
            type="submit"
            variant="destructive"
          >
            Delete my account
          </Button>
        </form>
      </Form>
      <ConfirmDialog
        cancelLabel="Keep account"
        confirmLabel="Delete account"
        confirmVariant="destructive"
        description="Your personal details are erased and you are signed out everywhere. This cannot be undone."
        isPending={mutation.isPending}
        onCancel={() => setOpen(false)}
        onConfirm={erase}
        open={open}
        title="Delete account?"
      />
    </>
  );
}
