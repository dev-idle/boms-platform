import type { ReactNode } from "react";

import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { cn } from "@/lib/utils";

type DashboardAccountProfileLayoutProps = {
  children: ReactNode;
  description: string;
  title: string;
};

/** Shared profile page chrome for internal dashboard roles. */
export function DashboardAccountProfileLayout({
  children,
  description,
  title,
}: DashboardAccountProfileLayoutProps) {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader description={description} title={title} />
      <div className="dashboard-page-body">{children}</div>
    </div>
  );
}

type DashboardProfileSectionProps = {
  children: ReactNode;
  description?: string;
  id: string;
  title: string;
  variant?: "default" | "flush-table";
};

export function DashboardProfileSection({
  children,
  description,
  id,
  title,
  variant = "default",
}: DashboardProfileSectionProps) {
  return (
    <section
      aria-labelledby={id}
      className={cn(
        "dashboard-profile-section",
        variant === "flush-table" && "dashboard-profile-section--flush-table",
      )}
    >
      <header className="dashboard-profile-section-header">
        <h2 className="dashboard-profile-section-title" id={id}>
          {title}
        </h2>
        {description ? (
          <p className="dashboard-profile-section-desc">{description}</p>
        ) : null}
      </header>
      <div className="dashboard-profile-section-body">{children}</div>
    </section>
  );
}
