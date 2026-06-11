"use client";

import { cn } from "@/lib/utils";

import type { AdminUser } from "../schemas";
import { AdminTableActionUnavailable } from "./admin-table-action-unavailable";

type AdminUserSessionActionProps = {
  className?: string;
  currentUserId?: string;
  onRevokeSessions: (user: AdminUser) => void;
  user: AdminUser;
};

/** Shield-off — admin session invalidation; distinct from sidebar sign-out. */
function RevokeSessionsIcon({ className }: { className?: string }) {
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
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10" />
      <path d="m4.5 4.5 15 15" />
    </svg>
  );
}

export function AdminUserSessionAction({
  className,
  currentUserId,
  onRevokeSessions,
  user,
}: AdminUserSessionActionProps) {
  const isSelf = currentUserId === user.id;

  if (isSelf) {
    return <AdminTableActionUnavailable />;
  }

  return (
    <button
      aria-label={`Revoke sessions for ${user.email}`}
      className={cn("db-table-action db-table-action--revoke", className)}
      onClick={() => onRevokeSessions(user)}
      title="Revoke sessions"
      type="button"
    >
      <RevokeSessionsIcon className="db-table-action-icon" />
    </button>
  );
}
