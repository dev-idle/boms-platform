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

  return (
    <div className="storefront-account-danger">
      <p className="storefront-account-danger__copy">
        This action disables your account and signs you out immediately. This
        cannot be undone from the app.
      </p>
      <Button
        disabled={!canDelete || mutation.isPending}
        onClick={() => setOpen(true)}
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
