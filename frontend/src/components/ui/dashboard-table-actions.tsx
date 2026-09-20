import type { ReactNode } from "react";
import Link from "next/link";

import { cn } from "@/lib/utils";

type DashboardTableEditLinkProps = {
  className?: string;
  href: string;
  /** Accessible name — carries the row identity (e.g. "Edit Croissant"). */
  label?: string;
  /** Visible word. */
  text?: string;
};

type DashboardTableDeleteButtonProps = {
  className?: string;
  disabled?: boolean;
  label?: string;
  onClick: () => void;
  text?: string;
};

/** Text row actions — shared across manager, staff, and admin tables. */
export function DashboardTableRowActions({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return <div className={cn("db-table-row-actions", className)}>{children}</div>;
}

export function DashboardTableEditLink({
  className,
  href,
  label = "Edit",
  text = "Edit",
}: DashboardTableEditLinkProps) {
  return (
    <Link
      aria-label={label}
      className={cn("db-table-action db-table-action--edit", className)}
      href={href}
    >
      {text}
    </Link>
  );
}

export function DashboardTableDeleteButton({
  className,
  disabled = false,
  label = "Delete",
  onClick,
  text = "Delete",
}: DashboardTableDeleteButtonProps) {
  return (
    <button
      aria-label={label}
      className={cn("db-table-action db-table-action--delete", className)}
      disabled={disabled}
      onClick={onClick}
      type="button"
    >
      {text}
    </button>
  );
}
