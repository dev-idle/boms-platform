"use client";

import * as React from "react";

import { cn } from "@/lib/utils";

import { Input, type InputProps } from "./input";

type IntegerFieldInputProps = Omit<
  InputProps,
  "type" | "value" | "onChange" | "onBlur" | "inputMode"
> & {
  onBlur?: () => void;
  onChange: (value: number) => void;
  value: number;
};

function parseIntegerDraft(raw: string, fallback: number, min: number): number {
  const trimmed = raw.trim();
  if (trimmed === "") {
    return min;
  }

  const parsed = Number.parseInt(trimmed, 10);
  if (Number.isNaN(parsed)) {
    return fallback;
  }

  return Math.max(min, parsed);
}

/** Dashboard integer field — clears zero on focus; native spinners hidden via `field-chrome--integer`. */
export const IntegerFieldInput = React.forwardRef<
  HTMLInputElement,
  IntegerFieldInputProps
>(function IntegerFieldInput(
  {
    className,
    min = 0,
    onBlur,
    onChange,
    onFocus,
    step = 1,
    value,
    ...props
  },
  ref,
) {
  const resolvedMin = typeof min === "number" ? min : Number(min);
  const safeMin = Number.isFinite(resolvedMin) ? resolvedMin : 0;
  const [draft, setDraft] = React.useState<string | null>(null);
  const displayValue = draft ?? String(value);

  function commitDraft(nextDraft: string | null): void {
    const resolved = parseIntegerDraft(nextDraft ?? "", value, safeMin);
    onChange(resolved);
    setDraft(null);
  }

  return (
    <Input
      {...props}
      className={cn("field-chrome--integer", className)}
      inputMode="numeric"
      min={safeMin}
      onBlur={() => {
        commitDraft(draft);
        onBlur?.();
      }}
      onChange={(event) => {
        const next = event.target.value;
        setDraft(next);

        if (next === "" || next === "-") {
          return;
        }

        const parsed = Number.parseInt(next, 10);
        if (!Number.isNaN(parsed)) {
          onChange(Math.max(safeMin, parsed));
        }
      }}
      onFocus={(event) => {
        setDraft(value === 0 ? "" : String(value));
        onFocus?.(event);
      }}
      ref={ref}
      step={step}
      type="number"
      value={displayValue}
    />
  );
});
IntegerFieldInput.displayName = "IntegerFieldInput";

