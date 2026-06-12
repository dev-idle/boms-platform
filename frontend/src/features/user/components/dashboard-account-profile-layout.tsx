import type { ReactNode } from "react";

import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";

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
    <div className="dashboard-page-stack dashboard-account-profile-page">
      <DashboardPageHeader description={description} title={title} />
      <div className="dashboard-profile-stack">{children}</div>
    </div>
  );
}

type DashboardProfileSectionProps = {
  children: ReactNode;
  description?: string;
  id: string;
  title: string;
};

export function DashboardProfileSection({
  children,
  description,
  id,
  title,
}: DashboardProfileSectionProps) {
  return (
    <section aria-labelledby={id} className="dashboard-profile-section">
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
