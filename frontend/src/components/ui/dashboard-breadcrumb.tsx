import Link from "next/link";

import { cn } from "@/lib/utils";

export type DashboardBreadcrumbItem = {
  href?: string;
  label: string;
};

type DashboardBreadcrumbProps = {
  className?: string;
  items: readonly DashboardBreadcrumbItem[];
};

export function DashboardBreadcrumb({
  className,
  items,
}: DashboardBreadcrumbProps) {
  if (items.length === 0) {
    return null;
  }

  const current = items[items.length - 1];
  const parent = items.length > 1 ? items[items.length - 2] : undefined;

  return (
    <nav aria-label="Breadcrumb" className={cn("dashboard-breadcrumb", className)}>
      {parent?.href ? (
        <Link className="dashboard-breadcrumb-back" href={parent.href}>
          {parent.label}
        </Link>
      ) : null}
      {items.length > 1 ? (
        <span aria-hidden className="dashboard-breadcrumb-sep">
          /
        </span>
      ) : null}
      <span
        aria-current="page"
        className="dashboard-breadcrumb-current"
        title={current.label}
      >
        {current.label}
      </span>
    </nav>
  );
}
