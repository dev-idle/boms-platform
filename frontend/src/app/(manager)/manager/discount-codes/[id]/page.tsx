"use client";

import { useRouter } from "next/navigation";
import { use } from "react";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { DiscountCodeForm, useDiscountCode } from "@/features/manager";
import { managerDiscountCodesBreadcrumb } from "@/features/manager/lib/manager-breadcrumbs";
import { ROUTE } from "@/constants/routes";
import { DashboardProfileSection } from "@/features/user";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

type PageProps = {
  params: Promise<{ id: string }>;
};

export default function ManagerEditDiscountCodePage({ params }: PageProps) {
  const { id } = use(params);
  const router = useRouter();
  const discountCodeQuery = useDiscountCode(id);

  if (discountCodeQuery.isPending) {
    return <p className="text-sm text-muted">Loading discount code…</p>;
  }

  if (discountCodeQuery.isError) {
    return <p className="text-sm text-error">Failed to load discount code.</p>;
  }

  if (!discountCodeQuery.data) {
    return <p className="text-sm text-muted">Discount code not found.</p>;
  }

  return (
    <DashboardFormPage
      breadcrumbItems={managerDiscountCodesBreadcrumb(PAGE_TITLES.editDiscountCode)}
      description="Update promotion rules and active window."
      title={PAGE_TITLES.editDiscountCode}
    >
      <DashboardProfileSection id="manager-discount-code-form" title="Code details">
        <DiscountCodeForm
          discountCode={discountCodeQuery.data}
          mode="edit"
          onSuccess={() => router.push(ROUTE.manager.discountCodes)}
        />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
