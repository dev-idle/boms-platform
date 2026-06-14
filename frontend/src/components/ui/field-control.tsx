"use client";

import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

import { FormControl, useFormField } from "./form";
import { Label } from "./label";

type FieldControlProps = {
  children: ReactNode;
  className?: string;
  label: string;
  /** Renders a muted “(optional)” suffix on the field label. */
  optional?: boolean;
};

/** Label + control group — must be used inside `FormField` + `FormItem`. */
export function FieldControl({
  label,
  children,
  className,
  optional = false,
}: FieldControlProps) {
  const { error, formItemId } = useFormField();

  return (
    <div className={cn("field-control", className)}>
      <Label
        className={cn(error && "text-error")}
        htmlFor={formItemId}
      >
        {label}
        {optional ? (
          <span className="field-label-optional">(optional)</span>
        ) : null}
      </Label>
      <FormControl>{children}</FormControl>
    </div>
  );
}
