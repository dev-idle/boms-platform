"use client";

import { useRouter } from "next/navigation";

import { ROUTE } from "@/constants/routes";
import { ProductForm } from "@/features/manager";

export default function ManagerNewProductPage() {
  const router = useRouter();

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-page-title">
          New product
        </h1>
      </div>
      <ProductForm
        mode="create"
        onSuccess={() => router.push(ROUTE.manager.products)}
      />
    </div>
  );
}
