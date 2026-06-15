"use client";

import { useId, type ReactNode } from "react";

import { cn } from "@/lib/utils";

import { FormControl, useFormField } from "./form";
import { FormFieldHint } from "./form-field-hint";
import { Label } from "./label";

type FieldControlProps = {
  children: ReactNode;
  className?: string;
  /** Helper copy when label + control are not enough — see `FORM_FIELD_HINT`. */
  hint?: string;
  hintId?: string;
  label: string;
  /** Renders a muted “(optional)” suffix on the field label. */
  optional?: boolean;
  variant?: "default" | "switch";
};

/** Label + control group — must be used inside `FormField` + `FormItem`. */
export function FieldControl({
  children,
  className,
  hint,
  hintId,
  label,
  optional = false,
  variant = "default",
}: FieldControlProps) {
  const { error, formItemId } = useFormField();
  const generatedHintId = useId();
  const resolvedHintId = hintId ?? generatedHintId;
  const describedBy = hint ? resolvedHintId : undefined;

  if (variant === "switch") {
    return (
      <div className={cn("field-control field-control--toggle", className)}>
        <div
          className={cn("form-toggle-row", hint && "form-toggle-row--with-hint")}
        >
          <div className="form-toggle-row-copy">
            <Label
              className={cn("form-toggle-row-label", error && "text-error")}
              htmlFor={formItemId}
            >
              {label}
              {optional ? (
                <span className="field-label-optional">(optional)</span>
              ) : null}
            </Label>
            {hint ? <FormFieldHint id={resolvedHintId}>{hint}</FormFieldHint> : null}
          </div>
          <FormControl aria-describedby={describedBy}>{children}</FormControl>
        </div>
      </div>
    );
  }

  return (
    <div className={cn("field-control", className)}>
      <div className="field-control-label-block">
        <Label
          className={cn(error && "text-error")}
          htmlFor={formItemId}
        >
          {label}
          {optional ? (
            <span className="field-label-optional">(optional)</span>
          ) : null}
        </Label>
        {hint ? <FormFieldHint id={resolvedHintId}>{hint}</FormFieldHint> : null}
      </div>
      <FormControl aria-describedby={describedBy}>{children}</FormControl>
    </div>
  );
}
