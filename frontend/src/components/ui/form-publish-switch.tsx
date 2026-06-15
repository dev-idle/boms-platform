"use client";

import type { ComponentPropsWithoutRef } from "react";

import { cn } from "@/lib/utils";

type FormPublishSwitchProps = Omit<
  ComponentPropsWithoutRef<"button">,
  "aria-checked" | "children" | "onClick" | "role" | "type"
> & {
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
};

/** Dashboard toggle — pair with `FieldControl` variant="switch" (inline row). */
export function FormPublishSwitch({
  checked,
  className,
  disabled = false,
  onCheckedChange,
  ...buttonProps
}: FormPublishSwitchProps) {
  return (
    <button
      {...buttonProps}
      aria-checked={checked}
      className={cn(
        "form-publish-switch",
        checked && "form-publish-switch--on",
        className,
      )}
      disabled={disabled}
      onClick={() => onCheckedChange(!checked)}
      role="switch"
      type="button"
    >
      <span className="sr-only">{checked ? "On" : "Off"}</span>
      <span aria-hidden className="form-publish-switch-thumb" />
    </button>
  );
}
