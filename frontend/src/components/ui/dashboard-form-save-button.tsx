"use client";

import { Button, type ButtonProps } from "@/components/ui/button";

type DashboardFormSaveButtonProps = {
  idleLabel: string;
  isPending: boolean;
  pendingLabel: string;
  /** Primary advances the page task; a secondary submit (e.g. password) is outline. */
  variant?: Extract<ButtonProps["variant"], "default" | "outline">;
};

/** Dashboard form submit — enabled until the mutation runs; Zod validates on submit. */
export function DashboardFormSaveButton({
  idleLabel,
  isPending,
  pendingLabel,
  variant = "default",
}: DashboardFormSaveButtonProps) {
  return (
    <Button aria-busy={isPending} disabled={isPending} type="submit" variant={variant}>
      {isPending ? pendingLabel : idleLabel}
    </Button>
  );
}
