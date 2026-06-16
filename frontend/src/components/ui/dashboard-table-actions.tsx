import type { MouseEventHandler, ReactNode } from "react";
import Link from "next/link";

import {
  DashboardAddIcon,
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

type DashboardTableAddButtonProps = {
  className?: string;
  disabled?: boolean;
  label?: string;
  onClick: () => void;
  onMouseDown?: MouseEventHandler<HTMLButtonElement>;
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

export function DashboardTableAddButton({
  className,
  disabled = false,
  label = "Add",
  onClick,
  onMouseDown,
}: DashboardTableAddButtonProps) {
  return (
    <button
      aria-label={label}
      className={cn("db-table-action db-table-action--add", className)}
      disabled={disabled}
      onClick={onClick}
      onMouseDown={onMouseDown}
      title={label}
      type="button"
    >
      <DashboardAddIcon className="db-table-action-icon" />
    </button>
  );
}
