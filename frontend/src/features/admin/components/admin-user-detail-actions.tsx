"use client";

import { Fragment } from "react";

import { ASSIGNABLE_OPERATIONAL_ROLES } from "@/constants/roles";
import { cn } from "@/lib/utils";

import { isAdminAccountManagementLocked } from "../lib/account-management";
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
  // A greyed action says why it is greyed, and while one action runs the others
  // are greyed for the same reason.
  const busyReason = actionsPending ? "Another action is finishing…" : undefined;
  const showReset = canResetPassword(user);
  const accountLocked = isAdminAccountManagementLocked(user);

  // §"Inline action groups take a hairline between members": three uppercase text
  // actions in a row read as one run, so a 1px rule divides them. Building the list
  // first is what lets the separators land only BETWEEN rendered members.
  const actions = [
    !accountLocked ? (
        user.disabled ? (
          <InlineAction
            disabled={actionsPending}
            onClick={onEnable}
            pendingLabel={accountPending ? "Enabling…" : undefined}
            title={busyReason}
            tone="accent"
          >
            Enable account
          </InlineAction>
        ) : (
          <InlineAction
            disabled={actionsPending}
            onClick={onDisable}
            pendingLabel={accountPending ? "Disabling…" : undefined}
            title={busyReason}
            tone="danger"
          >
            Disable account
          </InlineAction>
        )
      ) : null,
    !accountLocked ? (
      <InlineAction
          disabled={actionsPending}
          onClick={onRevokeSessions}
          title={busyReason}
          tone="warning"
        >
          Revoke all sessions
      </InlineAction>
    ) : null,
    showReset ? (
      <InlineAction
          disabled={actionsPending || user.disabled}
          onClick={onResetPassword}
          pendingLabel={resetPending ? "Generating…" : undefined}
          title={
            user.disabled
              ? "Enable this account before resetting the password."
              : busyReason
          }
          tone="accent"
        >
        Reset password
      </InlineAction>
    ) : null,
  ].filter(Boolean);

  return (
    <div className="dashboard-inline-actions">
      {actions.map((action, index) => (
        <Fragment key={index}>
          {index > 0 ? (
            <span aria-hidden className="dashboard-inline-actions__sep" />
          ) : null}
          {action}
        </Fragment>
      ))}
    </div>
  );
}
