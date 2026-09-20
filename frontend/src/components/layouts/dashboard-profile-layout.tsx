import type { ReactNode } from "react";

import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { cn } from "@/lib/utils";

type DashboardAccountProfileLayoutProps = {
  children: ReactNode;
  description: string;
  title: string;
};

/** Archetype D — self-service profile: header, then a single 640px column of sections. */
export function DashboardAccountProfileLayout({
  children,
  description,
  title,
}: DashboardAccountProfileLayoutProps) {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description={description}
        eyebrow={DASHBOARD_PAGE_EYEBROW.account}
        title={title}
      />
      <div className="dashboard-profile-column">{children}</div>
    </div>
  );
}

type DashboardProfileSectionProps = {
  children: ReactNode;
  description?: string;
  id: string;
  title: string;
  /**
   * `card` — bordered panel (manager forms, order summaries).
   * `plain` — archetype C/D section: display heading, sections divided by a hairline.
   */
  variant?: "card" | "plain";
};

export function DashboardProfileSection({
  children,
  description,
  id,
  title,
  variant = "card",
}: DashboardProfileSectionProps) {
  return (
    <section
      aria-labelledby={id}
      className={cn(
        "dashboard-profile-section",
        variant === "plain" && "dashboard-profile-section--plain",
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
