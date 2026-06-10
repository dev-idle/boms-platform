"use client";

import { useRouter } from "next/navigation";
import { use } from "react";

import { ROUTE } from "@/constants/routes";
import { ProductForm, useProduct } from "@/features/manager";

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
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-page-title">
          Edit product
        </h1>
      </div>
      <ProductForm
        mode="edit"
        product={productQuery.data}
        onSuccess={() => router.push(ROUTE.manager.products)}
      />
    </div>
  );
}
