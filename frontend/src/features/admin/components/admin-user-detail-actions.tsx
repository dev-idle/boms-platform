"use client";

import { ASSIGNABLE_OPERATIONAL_ROLES } from "@/constants/roles";
import { cn } from "@/lib/utils";

import type { AdminUser } from "../schemas";

type AdminUserDetailActionsProps = {
  accountPending?: boolean;
  onDisable: () => void;
  onEnable: () => void;
  onResetPassword: () => void;
  onRevokeSessions: () => void;
  resetPending?: boolean;
  user: AdminUser;
};

function canResetPassword(user: AdminUser): boolean {
  return ASSIGNABLE_OPERATIONAL_ROLES.includes(user.role);
}

type InlineActionProps = {
  children: string;
  disabled?: boolean;
  onClick: () => void;
  pendingLabel?: string;
  title?: string;
  tone: "accent" | "danger" | "warning";
};

function InlineAction({
  children,
  disabled = false,
  onClick,
  pendingLabel,
  title,
  tone,
}: InlineActionProps) {
  return (
    <button
      className={cn("dashboard-inline-action", `dashboard-inline-action--${tone}`)}
      disabled={disabled}
      onClick={onClick}
      title={title}
      type="button"
    >
      {pendingLabel ?? children}
    </button>
  );
}

export function AdminUserDetailActions({
  accountPending = false,
  onDisable,
  onEnable,
  onRevokeSessions,
  onResetPassword,
  resetPending = false,
  user,
}: AdminUserDetailActionsProps) {
  const actionsPending = accountPending || resetPending;
  const showReset = canResetPassword(user);

  return (
    <div className="dashboard-inline-actions">
      {user.disabled ? (
        <InlineAction
          disabled={actionsPending}
          onClick={onEnable}
          pendingLabel={accountPending ? "Enabling…" : undefined}
          tone="accent"
        >
          Enable account
        </InlineAction>
      ) : (
        <InlineAction
          disabled={actionsPending}
          onClick={onDisable}
          pendingLabel={accountPending ? "Disabling…" : undefined}
          tone="danger"
        >
          Disable account
        </InlineAction>
      )}
      <InlineAction
        disabled={actionsPending}
        onClick={onRevokeSessions}
        tone="warning"
      >
        Revoke all sessions
      </InlineAction>
      {showReset ? (
        <InlineAction
          disabled={actionsPending || user.disabled}
          onClick={onResetPassword}
          pendingLabel={resetPending ? "Generating…" : undefined}
          title={
            user.disabled
              ? "Enable this account before resetting the password."
              : undefined
          }
          tone="accent"
        >
          Reset password
        </InlineAction>
      ) : null}
    </div>
  );
}
