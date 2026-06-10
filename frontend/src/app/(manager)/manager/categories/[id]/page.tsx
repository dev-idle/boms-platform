"use client";

import { useRouter } from "next/navigation";
import { use } from "react";

import { ROUTE } from "@/constants/routes";
import { CategoryForm, useCategory } from "@/features/manager";

type PageProps = {
  params: Promise<{ id: string }>;
};

export default function ManagerEditCategoryPage({ params }: PageProps) {
  const { id } = use(params);
  const router = useRouter();
  const categoryQuery = useCategory(id);

  if (categoryQuery.isPending) {
    return <p className="text-sm text-muted">Loading category…</p>;
  }

  if (categoryQuery.isError) {
    return <p className="text-sm text-error">Failed to load category.</p>;
  }

  if (!categoryQuery.data) {
    return <p className="text-sm text-muted">Category not found.</p>;
  }

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-medium text-ink">
          Edit category
        </h1>
      </div>
      <CategoryForm
        category={categoryQuery.data}
        mode="edit"
        onSuccess={() => router.push(ROUTE.manager.categories)}
      />
    </div>
  );
}
