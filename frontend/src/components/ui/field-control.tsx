import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

import { FormControl, FormLabel } from "./form";

type FieldControlProps = {
  children: ReactNode;
  className?: string;
  label: string;
  /** Renders a muted “(optional)” suffix on the field label. */
  optional?: boolean;
};

/** Label + control group — hover/focus highlights label and field border. */
export function FieldControl({
  label,
  children,
  className,
  optional = false,
}: FieldControlProps) {
  return (
    <div className={cn("field-control", className)}>
      <FormLabel>
        {label}
        {optional ? <span className="field-label-optional">(optional)</span> : null}
      </FormLabel>
      <FormControl>{children}</FormControl>
    </div>
  );
}
