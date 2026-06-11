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

  return (
    <nav aria-label="Breadcrumb" className={cn("dashboard-breadcrumb", className)}>
      <ol className="dashboard-breadcrumb-list">
        {items.map((item, index) => {
          const isLast = index === items.length - 1;
          const isLink = Boolean(item.href) && !isLast;

          return (
            <li className="dashboard-breadcrumb-item" key={`${item.label}-${index}`}>
              {index > 0 ? (
                <span aria-hidden className="dashboard-breadcrumb-sep">
                  /
                </span>
              ) : null}
              {isLink && item.href ? (
                <Link className="dashboard-breadcrumb-link" href={item.href}>
                  {item.label}
                </Link>
              ) : (
                <span
                  aria-current={isLast ? "page" : undefined}
                  className="dashboard-breadcrumb-current"
                  title={item.label}
                >
                  {item.label}
                </span>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
