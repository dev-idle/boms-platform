"use client";

import { useRouter } from "next/navigation";
import { use } from "react";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { DashboardLoadError } from "@/components/ui/dashboard-load-error";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import {
  DiscountCodeForm,
  managerDiscountCodesBreadcrumb,
  useDiscountCode,
} from "@/features/manager";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

type PageProps = {
  params: Promise<{ id: string }>;
};

export default function ManagerEditDiscountCodePage({ params }: PageProps) {
  const { id } = use(params);
  const router = useRouter();
  const discountCodeQuery = useDiscountCode(id);

  if (discountCodeQuery.isPending) {
    return (
      <InlineLoadingState />
    );
  }

  if (discountCodeQuery.isError) {
    return isApiError(discountCodeQuery.error) && discountCodeQuery.error.status === 404 ? (
      <DashboardLoadError message="Discount code not found." />
    ) : (
      <DashboardLoadError message="We could not load this discount code." onRetry={() => void discountCodeQuery.refetch()} />
    );
  }

  return (
    <DashboardFormPage
      breadcrumbItems={managerDiscountCodesBreadcrumb(discountCodeQuery.data.code)}
      description="Update promotion rules and active window."
      title={PAGE_TITLES.editDiscountCode}
    >
      <DashboardProfileSection
        id="manager-discount-code-form"
        title="Code details"
        variant="plain"
      >
        <DiscountCodeForm
          discountCode={discountCodeQuery.data}
          mode="edit"
          onSuccess={() => router.push(ROUTE.manager.discountCodes)}
        />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
