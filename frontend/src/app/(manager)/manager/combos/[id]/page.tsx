"use client";

import { useRouter } from "next/navigation";
import { use } from "react";

import { ROUTE } from "@/constants/routes";
import { ComboForm, useCombo } from "@/features/manager";

type PageProps = {
  params: Promise<{ id: string }>;
};

export default function ManagerEditComboPage({ params }: PageProps) {
  const { id } = use(params);
  const router = useRouter();
  const comboQuery = useCombo(id);

  if (comboQuery.isPending) {
    return <p className="text-sm text-zinc-500">Loading combo…</p>;
  }

  if (comboQuery.isError) {
    return <p className="text-sm text-red-600">Failed to load combo.</p>;
  }

  if (!comboQuery.data) {
    return <p className="text-sm text-zinc-500">Combo not found.</p>;
  }

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
          Edit combo
        </h1>
      </div>
      <ComboForm
        combo={comboQuery.data}
        mode="edit"
        onSuccess={() => router.push(ROUTE.manager.combos)}
      />
    </div>
  );
}
