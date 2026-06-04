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
      <div className="rounded-lg border border-zinc-200 p-4 dark:border-zinc-800">
        <h2 className="font-semibold text-zinc-900 dark:text-zinc-100">
          Revoke sessions
        </h2>
        <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-300">
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

      <div className="rounded-lg border border-red-200 bg-red-50 p-4 dark:border-red-900/60 dark:bg-red-950/40">
        <h2 className="font-semibold text-red-800 dark:text-red-300">
          Disable user
        </h2>
        <p className="mt-1 text-sm text-red-700 dark:text-red-200">
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
