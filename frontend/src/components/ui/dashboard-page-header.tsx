import type { ReactNode } from "react";

import {
  DashboardBreadcrumb,
  type DashboardBreadcrumbItem,
} from "@/components/ui/dashboard-breadcrumb";
import { cn } from "@/lib/utils";

type DashboardPageHeaderProps = {
  actions?: ReactNode;
  breadcrumbItems?: readonly DashboardBreadcrumbItem[];
  className?: string;
  description?: ReactNode;
  meta?: ReactNode;
  title: string;
};

/** Shared dashboard page title block — rhythm via `.dashboard-page-header` in globals.css. */
export function DashboardPageHeader({
  actions,
  breadcrumbItems,
  className,
  description,
  meta,
  title,
}: DashboardPageHeaderProps) {
  const hasToolbar = Boolean(actions) && !meta;
  const hasBreadcrumb = Boolean(breadcrumbItems && breadcrumbItems.length > 0);

  const body = (
    <>
      {hasBreadcrumb && breadcrumbItems ? (
        <DashboardBreadcrumb items={breadcrumbItems} />
      ) : null}
      <h1 className="text-page-title">{title}</h1>
      {description ? <p className="dashboard-page-lead">{description}</p> : null}
      {meta}
    </>
  );

  return (
    <header
      className={cn(
        "dashboard-page-header",
        hasToolbar && "dashboard-page-header--toolbar",
        className,
      )}
    >
      {hasToolbar ? <div className="dashboard-page-header-body">{body}</div> : body}
      {hasToolbar && actions ? (
        <div className="dashboard-page-header-actions">{actions}</div>
      ) : null}
    </header>
  );
}
