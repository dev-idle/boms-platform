"use client";

import { useWatch, type FieldValues, type UseFormReturn } from "react-hook-form";

import { Button } from "@/components/ui/button";

type DashboardFormSaveButtonProps<T extends FieldValues> = {
  areEqual: (current: T, baseline: T) => boolean;
  baseline: T;
  form: UseFormReturn<T>;
  idleLabel: string;
  isPending: boolean;
  pendingLabel: string;
};

/**
 * Dirty-gated submit isolated from field rows so `useWatch` does not re-render
 * Controllers on every keystroke (React Compiler + RHF controlled input safe).
 */
export function DashboardFormSaveButton<T extends FieldValues>({
  areEqual,
  baseline,
  form,
  idleLabel,
  isPending,
  pendingLabel,
}: DashboardFormSaveButtonProps<T>) {
  const current = useWatch({ control: form.control }) as T;
  const hasUnsavedChanges = !areEqual(current, baseline);

  return (
    <Button disabled={isPending || !hasUnsavedChanges} type="submit">
      {isPending ? pendingLabel : idleLabel}
    </Button>
  );
}
