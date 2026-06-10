"use client";

import { useRouter } from "next/navigation";

import { ROUTE } from "@/constants/routes";
import { DiscountCodeForm } from "@/features/manager";

export default function ManagerNewDiscountCodePage() {
  const router = useRouter();

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-medium text-ink">
          New discount code
        </h1>
      </div>
      <DiscountCodeForm
        mode="create"
        onSuccess={() => router.push(ROUTE.manager.discountCodes)}
      />
    </div>
  );
}
