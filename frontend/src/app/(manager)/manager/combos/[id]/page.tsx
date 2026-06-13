"use client";

import { useRouter } from "next/navigation";
import { use } from "react";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { ComboForm, useCombo } from "@/features/manager";
import { managerCombosBreadcrumb } from "@/features/manager/lib/manager-breadcrumbs";
import { ROUTE } from "@/constants/routes";
import { DashboardProfileSection } from "@/features/user";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

type PageProps = {
  params: Promise<{ id: string }>;
};

export default function ManagerEditComboPage({ params }: PageProps) {
  const { id } = use(params);
  const router = useRouter();
  const comboQuery = useCombo(id);

  if (comboQuery.isPending) {
    return <p className="text-sm text-muted">Loading combo…</p>;
  }

  if (comboQuery.isError) {
    return <p className="text-sm text-error">Failed to load combo.</p>;
  }

  if (!comboQuery.data) {
    return <p className="text-sm text-muted">Combo not found.</p>;
  }

  return (
    <DashboardFormPage
      breadcrumbItems={managerCombosBreadcrumb(PAGE_TITLES.editCombo)}
      description="Update bundle items, pricing, and availability window."
      title={PAGE_TITLES.editCombo}
    >
      <DashboardProfileSection id="manager-combo-form" title="Combo details">
        <ComboForm
          combo={comboQuery.data}
          mode="edit"
          onSuccess={() => router.push(ROUTE.manager.combos)}
        />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
