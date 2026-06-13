"use client";

import { useRouter } from "next/navigation";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { ComboForm } from "@/features/manager";
import { managerCombosBreadcrumb } from "@/features/manager/lib/manager-breadcrumbs";
import { ROUTE } from "@/constants/routes";
import { DashboardProfileSection } from "@/features/user";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

export default function ManagerNewComboPage() {
  const router = useRouter();

  return (
    <DashboardFormPage
      breadcrumbItems={managerCombosBreadcrumb(PAGE_TITLES.newCombo)}
      description="Bundle products with promotional pricing and a time window."
      title={PAGE_TITLES.newCombo}
    >
      <DashboardProfileSection id="manager-combo-form" title="Combo details">
        <ComboForm mode="create" onSuccess={() => router.push(ROUTE.manager.combos)} />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
