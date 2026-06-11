import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

import { FormControl, FormLabel } from "./form";

type FieldControlProps = {
  children: ReactNode;
  className?: string;
  label: string;
};

/** Label + control group — hover/focus highlights label and field border. */
export function FieldControl({ label, children, className }: FieldControlProps) {
  return (
    <div className={cn("field-control", className)}>
      <FormLabel>{label}</FormLabel>
      <FormControl>{children}</FormControl>
    </div>
  );
}
