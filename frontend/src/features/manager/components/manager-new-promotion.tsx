"use client";

import { useRouter } from "next/navigation";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

import { managerPromotionsBreadcrumb } from "../lib/manager-breadcrumbs";
import { PromotionForm } from "./promotion-form";

/** Writing a promotion; once it is sent, back to the list where it shows sending. */
export function ManagerNewPromotion() {
  const router = useRouter();

  return (
    <DashboardFormPage
      breadcrumbItems={managerPromotionsBreadcrumb(PAGE_TITLES.newPromotion)}
      description="Write the email once; it goes out when you confirm."
      title={PAGE_TITLES.newPromotion}
    >
      <DashboardProfileSection id="manager-promotion-form" title="Email" variant="plain">
        <PromotionForm onSent={() => router.push(ROUTE.manager.promotions)} />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
