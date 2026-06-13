"use client";

import { Button } from "@/components/ui/button";

type DashboardFormSaveButtonProps = {
  idleLabel: string;
  isPending: boolean;
  pendingLabel: string;
};

/** Dashboard form submit — enabled until the mutation runs; Zod validates on submit. */
export function DashboardFormSaveButton({
  idleLabel,
  isPending,
  pendingLabel,
}: DashboardFormSaveButtonProps) {
  return (
    <Button disabled={isPending} type="submit">
      {isPending ? pendingLabel : idleLabel}
    </Button>
  );
}
