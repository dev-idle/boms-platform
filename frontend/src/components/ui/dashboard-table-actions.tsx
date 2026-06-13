import type { ReactNode } from "react";
import Link from "next/link";

import {
  DashboardDeleteIcon,
  DashboardEditIcon,
} from "@/components/icons/dashboard-ui-icons";
import { cn } from "@/lib/utils";

type DashboardTableEditLinkProps = {
  className?: string;
  href: string;
  label?: string;
};

type DashboardTableDeleteButtonProps = {
  className?: string;
  disabled?: boolean;
  label?: string;
  onClick: () => void;
};

/** Icon row actions — shared across manager, staff, and admin tables. */
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
}: DashboardTableEditLinkProps) {
  return (
    <Link
      aria-label={label}
      className={cn("db-table-action db-table-action--edit", className)}
      href={href}
      title={label}
    >
      <DashboardEditIcon className="db-table-action-icon" />
    </Link>
  );
}

export function DashboardTableDeleteButton({
  className,
  disabled = false,
  label = "Delete",
  onClick,
}: DashboardTableDeleteButtonProps) {
  return (
    <button
      aria-label={label}
      className={cn("db-table-action db-table-action--delete", className)}
      disabled={disabled}
      onClick={onClick}
      title={label}
      type="button"
    >
      <DashboardDeleteIcon className="db-table-action-icon" />
    </button>
  );
}
