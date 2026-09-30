import type * as React from "react";

import { cn } from "@/lib/utils";

type CheckboxFieldProps = Omit<React.ComponentProps<"input">, "type" | "children"> & {
  /** The words beside the box; they may hold links. */
  children: React.ReactNode;
  labelClassName?: string;
};

/**
 * A square checkbox with its label: the whole line toggles it, and links in the
 * label still open.
 */
export function CheckboxField({ children, className, labelClassName, ...props }: CheckboxFieldProps) {
  return (
    <label className={cn("form-checkbox", labelClassName)}>
      <input className={cn("form-checkbox__input", className)} type="checkbox" {...props} />
      <span>{children}</span>
    </label>
  );
}
