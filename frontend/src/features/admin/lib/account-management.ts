import type { StatusPillVariant } from "@/components/ui/status-pill";
import { USER_ROLE, type UserRole } from "@/constants/roles";

/** Admin accounts cannot be disabled or have sessions revoked from the admin UI. */
export function isAdminAccountManagementLocked(user: { role: UserRole }): boolean {
  return user.role === USER_ROLE.admin;
}

type AdminAccountStatus = {
  label: string;
  variant: StatusPillVariant;
};

/**
 * Account state for the status column. "Protected" is a distinct state, not a
 * disabled Active — admin rows cannot be mutated at all.
 */
export function adminUserAccountStatus(user: {
  disabled: boolean;
  role: UserRole;
}): AdminAccountStatus {
  if (isAdminAccountManagementLocked(user)) {
    return { label: "Protected", variant: "completed" };
  }

  return user.disabled
    ? { label: "Disabled", variant: "cancelled" }
    : { label: "Active", variant: "ready" };
}

type RowActionTarget = {
  disabled: boolean;
  id: string;
  role: UserRole;
};

type AdminRowAction =
  | { kind: "live" }
  | { kind: "blocked"; reason: string };

export type AdminUserRowActions = {
  /** Slot 1 — the label follows account state; blocked keeps the "Disable" word. */
  account: AdminRowAction & { label: "Disable" | "Enable" };
  /** Slot 2 — revoke sessions. */
  sessions: AdminRowAction;
};

/**
 * Fixed action set for an Admin Users row: an action that
 * cannot apply is greyed with its reason, never removed, so every slot keeps a
 * stable position down the column.
 */
export function adminUserRowActions(
  user: RowActionTarget,
  currentUserId: string | undefined,
): AdminUserRowActions {
  if (user.id === currentUserId) {
    return {
      account: {
        kind: "blocked",
        label: "Disable",
        reason: "You cannot disable your own account.",
      },
      sessions: { kind: "blocked", reason: "You cannot revoke your own sessions." },
    };
  }

  if (isAdminAccountManagementLocked(user)) {
    return {
      account: {
        kind: "blocked",
        label: "Disable",
        reason: "Admin accounts cannot be disabled here.",
      },
      sessions: { kind: "blocked", reason: "Admin sessions cannot be revoked here." },
    };
  }

  if (user.disabled) {
    return {
      account: { kind: "live", label: "Enable" },
      sessions: {
        kind: "blocked",
        reason: "Disabling this account already revoked its sessions.",
      },
    };
  }

  return {
    account: { kind: "live", label: "Disable" },
    sessions: { kind: "live" },
  };
}
