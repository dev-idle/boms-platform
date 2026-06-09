import Link from "next/link";
import type { ReactNode } from "react";

import { LogoutButton } from "@/features/auth";

type OperationalRoleShellProps = {
  roleLabel: string;
  profileHref: string;
  passwordHref: string;
  children: ReactNode;
};

/** Shared chrome for staff / baker / manager account areas (gate wraps outside). */
export function OperationalRoleShell({
  roleLabel,
  profileHref,
  passwordHref,
  children,
}: OperationalRoleShellProps) {
  return (
    <>
      <header className="border-b border-border bg-surface">
        <nav className="mx-auto flex max-w-5xl items-center gap-6 px-6 py-4 text-sm font-medium text-muted">
          <span className="role-badge">{roleLabel}</span>
          <Link className="transition-colors duration-default ease-default hover:text-foreground" href={profileHref}>
            Profile
          </Link>
          <Link className="transition-colors duration-default ease-default hover:text-foreground" href={passwordHref}>
            Password
          </Link>
          <div className="ml-auto flex items-center gap-3">
            <LogoutButton />
          </div>
        </nav>
      </header>
      <div className="mx-auto max-w-5xl px-6 py-10">{children}</div>
    </>
  );
}
