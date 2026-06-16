"use client";

import { useRouter } from "next/navigation";
import { use } from "react";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { CategoryForm, managerCategoriesBreadcrumb, useCategory } from "@/features/manager";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

type PageProps = {
  params: Promise<{ id: string }>;
};

export default function ManagerEditCategoryPage({ params }: PageProps) {
  const { id } = use(params);
  const router = useRouter();
  const categoryQuery = useCategory(id);

  if (categoryQuery.isPending) {
    return <InlineLoadingState />;
  }

  if (categoryQuery.isError) {
    return <p className="text-sm text-error">Failed to load category.</p>;
  }

  if (!categoryQuery.data) {
    return <p className="text-sm text-muted">Category not found.</p>;
  }

  return (
    <DashboardFormPage
      breadcrumbItems={managerCategoriesBreadcrumb(PAGE_TITLES.editCategory)}
      description="Update how this category appears in the catalog."
      title={PAGE_TITLES.editCategory}
    >
      <DashboardProfileSection id="manager-category-form" title="Category details">
        <CategoryForm
          category={categoryQuery.data}
          mode="edit"
          onSuccess={() => router.push(ROUTE.manager.categories)}
        />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
