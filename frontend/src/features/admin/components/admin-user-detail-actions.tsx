"use client";

import { Button } from "@/components/ui/button";

import type { AdminUser } from "../schemas";

type AdminUserDetailActionsProps = {
  isPending?: boolean;
  onDisable: () => void;
  onEnable: () => void;
  onRevokeSessions: () => void;
  user: AdminUser;
};

export function AdminUserDetailActions({
  isPending = false,
  onDisable,
  onEnable,
  onRevokeSessions,
  user,
}: AdminUserDetailActionsProps) {
  return (
    <div className="admin-user-detail-actions">
      {user.disabled ? (
        <Button disabled={isPending} onClick={onEnable} size="sm" type="button">
          {isPending ? "Enabling…" : "Enable account"}
        </Button>
      ) : (
        <Button
          disabled={isPending}
          onClick={onDisable}
          size="sm"
          type="button"
          variant="destructive"
        >
          {isPending ? "Disabling…" : "Disable account"}
        </Button>
      )}
      <Button
        disabled={isPending}
        onClick={onRevokeSessions}
        size="sm"
        type="button"
        variant="outline"
      >
        Revoke all sessions
      </Button>
    </div>
  );
}
