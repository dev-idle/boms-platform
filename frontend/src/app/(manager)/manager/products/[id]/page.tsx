"use client";

import { useRouter } from "next/navigation";
import { use } from "react";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { ProductForm, useProduct } from "@/features/manager";
import { managerProductsBreadcrumb } from "@/features/manager/lib/manager-breadcrumbs";
import { ROUTE } from "@/constants/routes";
import { DashboardProfileSection } from "@/features/user";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

type PageProps = {
  params: Promise<{ id: string }>;
};

export default function ManagerEditProductPage({ params }: PageProps) {
  const { id } = use(params);
  const router = useRouter();
  const productQuery = useProduct(id);

  if (productQuery.isPending) {
    return <p className="text-sm text-muted">Loading product…</p>;
  }

  if (productQuery.isError) {
    return <p className="text-sm text-error">Failed to load product.</p>;
  }

  if (!productQuery.data) {
    return <p className="text-sm text-muted">Product not found.</p>;
  }

  return (
    <DashboardFormPage
      breadcrumbItems={managerProductsBreadcrumb(PAGE_TITLES.editProduct)}
      description="Update catalog fields shown on the storefront."
      title={PAGE_TITLES.editProduct}
    >
      <DashboardProfileSection id="manager-product-form" title="Product details">
        <ProductForm
          mode="edit"
          product={productQuery.data}
          onSuccess={() => router.push(ROUTE.manager.products)}
        />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
