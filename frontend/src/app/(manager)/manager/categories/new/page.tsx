"use client";

import { useRouter } from "next/navigation";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { CategoryForm } from "@/features/manager";
import { managerCategoriesBreadcrumb } from "@/features/manager/lib/manager-breadcrumbs";
import { ROUTE } from "@/constants/routes";
import { DashboardProfileSection } from "@/features/user";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

export default function ManagerNewCategoryPage() {
  const router = useRouter();

  return (
    <DashboardFormPage
      breadcrumbItems={managerCategoriesBreadcrumb(PAGE_TITLES.newCategory)}
      description="Group products for customer browsing."
      title={PAGE_TITLES.newCategory}
    >
      <DashboardProfileSection id="manager-category-form" title="Category details">
        <CategoryForm
          mode="create"
          onSuccess={() => router.push(ROUTE.manager.categories)}
        />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
