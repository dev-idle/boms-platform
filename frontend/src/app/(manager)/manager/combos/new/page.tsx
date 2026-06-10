"use client";

import { useRouter } from "next/navigation";

import { ROUTE } from "@/constants/routes";
import { ComboForm } from "@/features/manager";

export default function ManagerNewComboPage() {
  const router = useRouter();

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-medium text-ink">
          New combo
        </h1>
      </div>
      <ComboForm
        mode="create"
        onSuccess={() => router.push(ROUTE.manager.combos)}
      />
    </div>
  );
}
