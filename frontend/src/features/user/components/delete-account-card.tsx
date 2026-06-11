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
    <div className="rounded-lg border border-error/30 bg-error-bg p-4 rounded-card">
      <h2 className="text-lg font-medium text-red-800">
        Delete account
      </h2>
      <p className="mt-2 text-sm text-red-700">
        This action disables your account and signs you out immediately.
      </p>
      <Button
        className="mt-4"
        disabled={!canDelete || mutation.isPending}
        onClick={() => setOpen(true)}
        type="button"
        variant="outline"
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
