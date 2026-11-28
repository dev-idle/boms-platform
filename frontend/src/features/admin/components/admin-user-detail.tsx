"use client";

import { useState } from "react";
import { toast } from "sonner";

import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { DashboardLoadError } from "@/components/ui/dashboard-load-error";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { useAuthStore } from "@/stores/auth-store";

import {
  useDisable,
  useEnable,
  useResetPassword,
  useRevokeSessions,
  useUserDetail,
} from "../hooks";
import { accountActionErrorMessage } from "../lib/account-action-error";
import { isAdminAccountManagementLocked } from "../lib/account-management";

import { AdminUserAccountStatus } from "./admin-user-account-status";
import { AdminUserActivityLog } from "./admin-user-activity-log";
import { AdminUserDetailActions } from "./admin-user-detail-actions";
import { AdminUserDetailHeader } from "./admin-user-detail-header";
import { AdminUserDetailManagement } from "./admin-user-detail-management";
import { TempPasswordModal } from "./temp-password-modal";

type PendingAction = "disable" | "enable" | "revoke" | "reset";

export function AdminUserDetail({ userId }: { userId: string }) {
  const currentUserId = useAuthStore((state) => state.user?.id);
  const userQuery = useUserDetail(userId);
  const resetPassword = useResetPassword();
  const disableUser = useDisable();
  const enableUser = useEnable();
  const revokeSessions = useRevokeSessions();
  const [pendingAction, setPendingAction] = useState<PendingAction | null>(null);

  const user = userQuery.data;
  const accountPending =
    disableUser.isPending || enableUser.isPending || revokeSessions.isPending;

  function clearPendingAction(): void {
    setPendingAction(null);
  }

  function handleActionError(error: unknown, fallback: string): void {
    toast.error(accountActionErrorMessage(error, fallback));
  }

  return (
    <>
      {userQuery.isPending ? (
        <InlineLoadingState />
      ) : !user ? (
        <DashboardLoadError message="We could not load this user." onRetry={() => void userQuery.refetch()} />
      ) : (
        <div className="dashboard-page-stack">
          <AdminUserDetailHeader
            actions={
              currentUserId !== user.id && !isAdminAccountManagementLocked(user) ? (
                <AdminUserDetailActions
                  accountPending={accountPending}
                  onDisable={() => setPendingAction("disable")}
                  onEnable={() => setPendingAction("enable")}
                  onResetPassword={() => setPendingAction("reset")}
                  onRevokeSessions={() => setPendingAction("revoke")}
                  resetPending={resetPassword.isPending}
                  user={user}
                />
              ) : undefined
            }
            user={user}
          />

          {/* Archetype C: work on the left, read-only facts in the right rail. */}
          <div className="dashboard-detail-grid">
            <div className="dashboard-detail-grid__main">
              <AdminUserDetailManagement user={user} userId={userId} />
              <AdminUserActivityLog userId={userId} />
            </div>
            <AdminUserAccountStatus user={user} />
          </div>
        </div>
      )}

      <TempPasswordModal
        data={resetPassword.tempPasswordData}
        onClose={resetPassword.clearTempPasswordData}
        open={Boolean(resetPassword.tempPasswordData)}
      />

      <ConfirmDialog
        confirmLabel="Disable account"
        confirmVariant="destructive"
        description="The account can no longer sign in and its sessions end. You can enable it again."
        isPending={disableUser.isPending}
        onCancel={clearPendingAction}
        onConfirm={() => {
          disableUser.mutate(userId, {
            onSuccess: clearPendingAction,
            onError: (error) => {
              handleActionError(error, "We could not disable account. Please try again.");
              clearPendingAction();
            },
          });
        }}
        open={pendingAction === "disable"}
        title="Disable account?"
      />

      <ConfirmDialog
        confirmLabel="Enable account"
        description="This account will be restored and can sign in again."
        isPending={enableUser.isPending}
        onCancel={clearPendingAction}
        onConfirm={() => {
          enableUser.mutate(userId, {
            onSuccess: clearPendingAction,
            onError: (error) => {
              handleActionError(error, "We could not enable account. Please try again.");
              clearPendingAction();
            },
          });
        }}
        open={pendingAction === "enable"}
        title="Enable account?"
      />

      <ConfirmDialog
        confirmLabel="Revoke sessions"
        confirmVariant="warning"
        description="All active sessions for this user will be revoked."
        isPending={revokeSessions.isPending}
        onCancel={clearPendingAction}
        onConfirm={() => {
          revokeSessions.mutate(userId, {
            onSuccess: clearPendingAction,
            onError: (error) => {
              handleActionError(error, "We could not revoke sessions. Please try again.");
              clearPendingAction();
            },
          });
        }}
        open={pendingAction === "revoke"}
        title="Revoke all sessions?"
      />

      <ConfirmDialog
        confirmLabel="Reset password"
        description="A new temporary password will be generated and shown once. All active sessions will be revoked."
        isPending={resetPassword.isPending}
        onCancel={clearPendingAction}
        onConfirm={() =>
          resetPassword.mutate(userId, {
            onSuccess: () => clearPendingAction(),
            onError: (error) => {
              handleActionError(error, "We could not reset password. Please try again.");
              clearPendingAction();
            },
          })
        }
        open={pendingAction === "reset"}
        title="Reset password?"
      />
    </>
  );
}
