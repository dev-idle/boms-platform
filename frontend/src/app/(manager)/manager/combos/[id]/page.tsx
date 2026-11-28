"use client";

import { useRouter } from "next/navigation";
import { use } from "react";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { DashboardLoadError } from "@/components/ui/dashboard-load-error";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { ComboForm, managerCombosBreadcrumb, useCombo } from "@/features/manager";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

type PageProps = {
  params: Promise<{ id: string }>;
};

export default function ManagerEditComboPage({ params }: PageProps) {
  const { id } = use(params);
  const router = useRouter();
  const comboQuery = useCombo(id);

  if (comboQuery.isPending) {
    return <InlineLoadingState />;
  }

  if (comboQuery.isError) {
    return isApiError(comboQuery.error) && comboQuery.error.status === 404 ? (
      <DashboardLoadError message="Combo not found." />
    ) : (
      <DashboardLoadError message="We could not load this combo." onRetry={() => void comboQuery.refetch()} />
    );
  }

  return (
    <DashboardFormPage
      breadcrumbItems={managerCombosBreadcrumb(comboQuery.data.name)}
      description="Update bundle items, pricing, and availability window."
      title={PAGE_TITLES.editCombo}
    >
      <DashboardProfileSection
        id="manager-combo-form"
        title="Combo details"
        variant="plain"
      >
        <ComboForm
          combo={comboQuery.data}
          mode="edit"
          onSuccess={() => router.push(ROUTE.manager.combos)}
        />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
