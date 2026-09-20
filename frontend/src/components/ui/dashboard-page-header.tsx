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
  /** Micro-label above the title — the section this page belongs to. */
  eyebrow?: string;
  /** Renders the description opposite the title block. */
  leadAside?: boolean;
  meta?: ReactNode;
  title: string;
};

/** Shared dashboard page title block — rhythm via `.dashboard-page-header` in globals.css. */
export function DashboardPageHeader({
  actions,
  breadcrumbItems,
  className,
  description,
  eyebrow,
  leadAside = false,
  meta,
  title,
}: DashboardPageHeaderProps) {
  const hasAsideLead = leadAside && Boolean(description);
  const hasToolbar = (Boolean(actions) || hasAsideLead) && !meta;
  const hasBreadcrumb = Boolean(breadcrumbItems && breadcrumbItems.length > 0);

  const body = (
    <>
      {hasBreadcrumb && breadcrumbItems ? (
        <DashboardBreadcrumb items={breadcrumbItems} />
      ) : null}
      {eyebrow ? <p className="dashboard-page-eyebrow">{eyebrow}</p> : null}
      <h1 className="text-page-title">{title}</h1>
      {description && !hasAsideLead ? (
        <p className="dashboard-page-lead">{description}</p>
      ) : null}
      {meta}
    </>
  );

  return (
    <header
      className={cn(
        "dashboard-page-header",
        hasToolbar && "dashboard-page-header--toolbar",
        hasAsideLead && "dashboard-page-header--lead-aside",
        className,
      )}
    >
      {hasToolbar ? <div className="dashboard-page-header-body">{body}</div> : body}
      {hasAsideLead ? (
        <p className="dashboard-page-lead dashboard-page-lead--aside">
          {description}
        </p>
      ) : null}
      {hasToolbar && actions ? (
        <div className="dashboard-page-header-actions">{actions}</div>
      ) : null}
    </header>
  );
}
