"use client";

import { cn } from "@/lib/utils";

import type { AdminUser } from "../schemas";
import { AdminTableActionUnavailable } from "./admin-table-action-unavailable";

type AdminUserAccountStatusToggleProps = {
  currentUserId?: string;
  onToggleAccount: (user: AdminUser) => void;
  user: AdminUser;
};

function ActiveStatusIcon({ className }: { className?: string }) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      viewBox="0 0 24 24"
    >
      <path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10Z" />
      <path d="m9 12 2 2 4-4" />
    </svg>
  );
}

function DisabledStatusIcon({ className }: { className?: string }) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      viewBox="0 0 24 24"
    >
      <path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10Z" />
      <path d="m9 9 6 6" />
      <path d="m15 9-6 6" />
    </svg>
  );
}

export function AdminUserAccountStatusToggle({
  currentUserId,
  onToggleAccount,
  user,
}: AdminUserAccountStatusToggleProps) {
  const isSelf = currentUserId === user.id;
  const isActive = !user.disabled;

  if (isSelf) {
    return <AdminTableActionUnavailable />;
  }

  return (
    <button
      aria-label={
        isActive
          ? `Disable account for ${user.email}`
          : `Enable account for ${user.email}`
      }
      className={cn(
        "db-table-status-toggle",
        isActive
          ? "db-table-status-toggle--active"
          : "db-table-status-toggle--disabled",
      )}
      onClick={() => onToggleAccount(user)}
      title={isActive ? "Disable account" : "Enable account"}
      type="button"
    >
      {isActive ? (
        <ActiveStatusIcon className="db-table-action-icon" />
      ) : (
        <DisabledStatusIcon className="db-table-action-icon" />
      )}
    </button>
  );
}
