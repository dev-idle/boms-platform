"use client";

import { useRouter } from "next/navigation";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { ProductForm } from "@/features/manager";
import { managerProductsBreadcrumb } from "@/features/manager/lib/manager-breadcrumbs";
import { ROUTE } from "@/constants/routes";
import { DashboardProfileSection } from "@/features/user";
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
