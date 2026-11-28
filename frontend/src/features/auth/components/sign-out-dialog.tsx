"use client";

import { ConfirmDialog } from "@/components/ui/confirm-dialog";

import { useLogout } from "../hooks";

type SignOutDialogProps = {
  open: boolean;
  onClose: () => void;
};

/** Asks before ending the session, so a slip on Sign out does not cost a sign-in. */
export function SignOutDialog({ open, onClose }: SignOutDialogProps) {
  const logout = useLogout();

  return (
    <ConfirmDialog
      confirmLabel="Sign out"
      confirmVariant="destructive"
      description="You will need your email and password to sign in again."
      isPending={logout.isPending}
      onCancel={onClose}
      onConfirm={() => logout.mutate()}
      open={open}
      title="Sign out?"
    />
  );
}
