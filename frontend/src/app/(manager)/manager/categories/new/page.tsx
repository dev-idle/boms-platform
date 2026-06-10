"use client";

import { useRouter } from "next/navigation";

import { ROUTE } from "@/constants/routes";
import { CategoryForm } from "@/features/manager";

export default function ManagerNewCategoryPage() {
  const router = useRouter();

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-medium text-ink">
          New category
        </h1>
      </div>
      <CategoryForm
        mode="create"
        onSuccess={() => router.push(ROUTE.manager.categories)}
      />
    </div>
  );
}
