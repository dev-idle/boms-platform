"use client";

import { useRouter } from "next/navigation";
import { use } from "react";

import { ROUTE } from "@/constants/routes";
import { DiscountCodeForm, useDiscountCode } from "@/features/manager";

type PageProps = {
  params: Promise<{ id: string }>;
};

export default function ManagerEditDiscountCodePage({ params }: PageProps) {
  const { id } = use(params);
  const router = useRouter();
  const discountCodeQuery = useDiscountCode(id);

  if (discountCodeQuery.isPending) {
    return <p className="text-sm text-zinc-500">Loading discount code…</p>;
  }

  if (discountCodeQuery.isError) {
    return <p className="text-sm text-red-600">Failed to load discount code.</p>;
  }

  if (!discountCodeQuery.data) {
    return <p className="text-sm text-zinc-500">Discount code not found.</p>;
  }

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
          Edit discount code
        </h1>
      </div>
      <DiscountCodeForm
        discountCode={discountCodeQuery.data}
        mode="edit"
        onSuccess={() => router.push(ROUTE.manager.discountCodes)}
      />
    </div>
  );
}
