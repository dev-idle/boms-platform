"use client";

import { useRouter } from "next/navigation";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { ProductForm, managerProductsBreadcrumb } from "@/features/manager";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

export default function ManagerNewProductPage() {
  const router = useRouter();

  return (
    <DashboardFormPage
      breadcrumbItems={managerProductsBreadcrumb(PAGE_TITLES.newProduct)}
      description="Add an item to the customer storefront catalog."
      title={PAGE_TITLES.newProduct}
    >
      <DashboardProfileSection id="manager-product-form" title="Product details">
        <ProductForm
          mode="create"
          onSuccess={() => router.push(ROUTE.manager.products)}
        />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
