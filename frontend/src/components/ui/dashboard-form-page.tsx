import type { ReactNode } from "react";

import type { DashboardBreadcrumbItem } from "@/components/ui/dashboard-breadcrumb";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";

type DashboardFormPageProps = {
  breadcrumbItems?: readonly DashboardBreadcrumbItem[];
  children: ReactNode;
  description?: string;
  title: string;
};

/** Archetype D shell for dashboard create/edit routes — one 640px column. */
export function DashboardFormPage({
  breadcrumbItems,
  children,
  description,
  title,
}: DashboardFormPageProps) {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        breadcrumbItems={breadcrumbItems}
        description={description}
        title={title}
      />
      <div className="dashboard-profile-column">{children}</div>
    </div>
  );
}
