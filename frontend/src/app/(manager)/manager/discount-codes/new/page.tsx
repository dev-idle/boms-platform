"use client";

import { useRouter } from "next/navigation";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { DiscountCodeForm, managerDiscountCodesBreadcrumb } from "@/features/manager";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

export default function ManagerNewDiscountCodePage() {
  const router = useRouter();

  return (
    <DashboardFormPage
      breadcrumbItems={managerDiscountCodesBreadcrumb(PAGE_TITLES.newDiscountCode)}
      description="Create a promotion code validated at checkout."
      title={PAGE_TITLES.newDiscountCode}
    >
      <DashboardProfileSection id="manager-discount-code-form" title="Code details">
        <DiscountCodeForm
          mode="create"
          onSuccess={() => router.push(ROUTE.manager.discountCodes)}
        />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
