"use client";

import { useRouter } from "next/navigation";
import { use } from "react";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { DashboardLoadError } from "@/components/ui/dashboard-load-error";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { ProductForm, managerProductsBreadcrumb, useProduct } from "@/features/manager";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

type PageProps = {
  params: Promise<{ id: string }>;
};

export default function ManagerEditProductPage({ params }: PageProps) {
  const { id } = use(params);
  const router = useRouter();
  const productQuery = useProduct(id);

  if (productQuery.isPending) {
    return <InlineLoadingState />;
  }

  if (productQuery.isError) {
    return isApiError(productQuery.error) && productQuery.error.status === 404 ? (
      <DashboardLoadError message="Product not found." />
    ) : (
      <DashboardLoadError message="We could not load this product." onRetry={() => void productQuery.refetch()} />
    );
  }

  return (
    <DashboardFormPage
      breadcrumbItems={managerProductsBreadcrumb(productQuery.data.name)}
      description="Update catalog fields shown on the storefront."
      title={PAGE_TITLES.editProduct}
    >
      <DashboardProfileSection
        id="manager-product-form"
        title="Product details"
        variant="plain"
      >
        <ProductForm
          mode="edit"
          product={productQuery.data}
          onSuccess={() => router.push(ROUTE.manager.products)}
        />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
