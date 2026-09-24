"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";

import { useDeleteAccount, useMe } from "../hooks";

export function DeleteAccountCard() {
  const mutation = useDeleteAccount();
  const me = useMe();
  const [open, setOpen] = useState(false);

  const canDelete = me.data?.role === "customer";
  // A greyed action says why it is greyed (02-COMPONENTS §4): still loading, or
  // not a customer account.
  const disabledReason = me.isPending
    ? "Loading your account…"
    : canDelete
      ? undefined
      : "Only customer accounts can be deleted here.";

  return (
    <div className="storefront-account-danger">
      <p className="storefront-account-danger__copy">
        This action disables your account and signs you out immediately. This
        cannot be undone from the app.
      </p>
      <Button
        aria-busy={me.isPending || undefined}
        disabled={!canDelete || mutation.isPending}
        onClick={() => setOpen(true)}
        title={disabledReason}
        type="button"
        variant="destructive"
      >
        Delete my account
      </Button>
      <ConfirmDialog
        cancelLabel="Keep account"
        confirmLabel="Delete account"
        confirmVariant="destructive"
        description="This cannot be undone from the app."
        isPending={mutation.isPending}
        onCancel={() => setOpen(false)}
        onConfirm={() => mutation.mutate()}
        open={open}
        title="Delete account?"
      />
    </div>
  );
}
