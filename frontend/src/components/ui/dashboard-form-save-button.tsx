"use client";

import { Button, type ButtonProps } from "@/components/ui/button";

type DashboardFormSaveButtonProps = {
  idleLabel: string;
  isPending: boolean;
  pendingLabel: string;
  /** Primary advances the page task; a secondary submit (e.g. password) is outline. */
  variant?: Extract<ButtonProps["variant"], "default" | "outline">;
  /** When set, the form cannot be submitted: greyed in place, with the reason as its title. */
  blockedReason?: string;
};

/** Dashboard form submit — disabled while the mutation runs or a blockedReason holds; Zod validates on submit. */
export function DashboardFormSaveButton({
  idleLabel,
  isPending,
  pendingLabel,
  variant = "default",
  blockedReason,
}: DashboardFormSaveButtonProps) {
  return (
    <Button
      aria-busy={isPending}
      disabled={isPending || Boolean(blockedReason)}
      title={blockedReason}
      type="submit"
      variant={variant}
    >
      {isPending ? pendingLabel : idleLabel}
    </Button>
  );
}
