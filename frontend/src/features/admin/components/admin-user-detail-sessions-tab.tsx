"use client";

import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { isApiError } from "@/lib/errors";

import { useDisable, useRevokeSessions } from "../hooks";

type AdminUserDetailSessionsTabProps = {
  userId: string;
};

export function AdminUserDetailSessionsTab({
  userId,
}: AdminUserDetailSessionsTabProps) {
  const [confirmDisable, setConfirmDisable] = useState(false);
  const [confirmRevoke, setConfirmRevoke] = useState(false);

  const disableUser = useDisable();
  const revokeSessions = useRevokeSessions();

  return (
    <div className="max-w-2xl space-y-4">
      <div className="rounded-lg border border-border p-4 rounded-card border bg-surface">
        <h2 className="font-medium text-ink">
          Revoke sessions
        </h2>
        <p className="mt-1 text-sm text-ink-2">
          Force this user to sign in again on all devices.
        </p>
        <Button
          className="mt-4"
          onClick={() => setConfirmRevoke(true)}
          type="button"
          variant="outline"
        >
          Revoke sessions
        </Button>
      </div>

      <div className="rounded-lg border border-error/30 bg-error-bg p-4 rounded-card">
        <h2 className="font-medium text-red-800">
          Disable user
        </h2>
        <p className="mt-1 text-sm text-red-700">
          This action soft-deletes the account.
        </p>
        <Button
          className="mt-4"
          onClick={() => setConfirmDisable(true)}
          type="button"
          variant="outline"
        >
          Disable account
        </Button>
      </div>

      <ConfirmDialog
        confirmLabel="Disable user"
        description="The user account will be disabled immediately."
        isPending={disableUser.isPending}
        onCancel={() => setConfirmDisable(false)}
        onConfirm={() =>
          disableUser.mutate(userId, {
            onSuccess: () => {
              setConfirmDisable(false);
            },
            onError: (error) => {
              if (!isApiError(error)) {
                toast.error("Failed to disable user");
                return;
              }
              if (error.isCannotModifySelf()) {
                toast.error("You cannot disable your own account.");
                return;
              }
              toast.error(error.message);
            },
          })
        }
        open={confirmDisable}
        title="Disable this user?"
      />

      <ConfirmDialog
        confirmLabel="Revoke sessions"
        description="All active sessions for this user will be revoked."
        isPending={revokeSessions.isPending}
        onCancel={() => setConfirmRevoke(false)}
        onConfirm={() =>
          revokeSessions.mutate(userId, {
            onSuccess: () => {
              setConfirmRevoke(false);
            },
            onError: (error) => {
              if (!isApiError(error)) {
                toast.error("Failed to revoke sessions");
                return;
              }
              if (error.isCannotModifySelf()) {
                toast.error("Use logout to end your own sessions.");
                return;
              }
              toast.error(error.message);
            },
          })
        }
        open={confirmRevoke}
        title="Revoke all sessions?"
      />
    </div>
  );
}
