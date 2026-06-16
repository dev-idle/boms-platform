"use client";

import * as React from "react";

import { cn } from "@/lib/utils";

import { Input, type InputProps } from "./input";

type IntegerFieldInputProps = Omit<
  InputProps,
  "type" | "value" | "onChange" | "onBlur" | "inputMode"
> & {
  /** Empty blur commits `undefined` instead of clamping to `min`. */
  allowEmpty?: boolean;
  max?: number;
  onBlur?: () => void;
  onChange: (value: number | undefined) => void;
  value: number | undefined;
};

export function parseIntegerFieldDraft(
  raw: string,
  fallback: number,
  min: number,
  max?: number,
): number {
  const trimmed = raw.trim();
  if (trimmed === "" || trimmed === "-") {
    return min;
  }

  const parsed = Number.parseInt(trimmed, 10);
  if (Number.isNaN(parsed)) {
    return fallback;
  }

  return clampInteger(parsed, min, max);
}

function clampInteger(value: number, min: number, max?: number): number {
  const effectiveMax =
    max !== undefined && max >= min ? max : undefined;
  let resolved = Math.max(min, value);
  if (effectiveMax !== undefined) {
    resolved = Math.min(effectiveMax, resolved);
  }
  return resolved;
}

function resolveIntegerBounds(
  min: number | string | undefined,
  max?: number,
): { safeMin: number; safeMax: number | undefined } {
  const resolvedMin = typeof min === "number" ? min : Number(min);
  const safeMin = Number.isFinite(resolvedMin) ? resolvedMin : 0;
  const resolvedMax = max === undefined ? undefined : Number(max);
  const safeMax =
    resolvedMax !== undefined && Number.isFinite(resolvedMax)
      ? resolvedMax
      : undefined;
  return { safeMin, safeMax };
}

/** Dashboard integer field — clears zero on focus; native spinners hidden via `field-chrome--integer`. */
export const IntegerFieldInput = React.forwardRef<
  HTMLInputElement,
  IntegerFieldInputProps
>(function IntegerFieldInput(
  {
    allowEmpty = false,
    className,
    max,
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
  const { safeMin, safeMax } = resolveIntegerBounds(min, max);
  const [draft, setDraft] = React.useState<string | null>(null);
  const displayValue =
    draft ?? (allowEmpty && value === undefined ? "" : String(value ?? 0));

  function commitDraft(raw: string): void {
    const trimmed = raw.trim();
    if (allowEmpty && (trimmed === "" || trimmed === "-")) {
      onChange(undefined);
      setDraft(null);
      return;
    }

    const fallback = value ?? safeMin;
    const resolved = parseIntegerFieldDraft(raw, fallback, safeMin, safeMax);
    onChange(resolved);
    setDraft(null);
  }

  return (
    <Input
      {...props}
      className={cn("field-chrome--integer", className)}
      inputMode="numeric"
      max={safeMax}
      min={safeMin}
      onBlur={(event) => {
        commitDraft(event.currentTarget.value);
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
          onChange(clampInteger(parsed, safeMin, safeMax));
        }
      }}
      onFocus={(event) => {
        if (allowEmpty && value === undefined) {
          setDraft("");
        } else {
          setDraft(value === 0 ? "" : String(value ?? 0));
        }
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

type OptionalIntegerFieldInputProps = Omit<
  InputProps,
  "type" | "value" | "onChange" | "onBlur" | "inputMode"
> & {
  max?: number;
  onBlur?: () => void;
  onChange: (value: number | null) => void;
  value: number | null;
};

/** Optional integer — empty input maps to `null` (dashboard optional limits). */
export const OptionalIntegerFieldInput = React.forwardRef<
  HTMLInputElement,
  OptionalIntegerFieldInputProps
>(function OptionalIntegerFieldInput(
  {
    className,
    max,
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
  const { safeMin, safeMax } = resolveIntegerBounds(min, max);
  const [draft, setDraft] = React.useState<string | null>(null);
  const displayValue = draft ?? (value === null ? "" : String(value));

  function commitDraft(raw: string): void {
    const trimmed = raw.trim();
    if (trimmed === "" || trimmed === "-") {
      onChange(null);
      setDraft(null);
      return;
    }

    const fallback = value ?? safeMin;
    onChange(parseIntegerFieldDraft(raw, fallback, safeMin, safeMax));
    setDraft(null);
  }

  return (
    <Input
      {...props}
      className={cn("field-chrome--integer", className)}
      inputMode="numeric"
      max={safeMax}
      min={safeMin}
      onBlur={(event) => {
        commitDraft(event.currentTarget.value);
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
          onChange(clampInteger(parsed, safeMin, safeMax));
        }
      }}
      onFocus={(event) => {
        setDraft(value === null ? "" : String(value));
        onFocus?.(event);
      }}
      ref={ref}
      step={step}
      type="number"
      value={displayValue}
    />
  );
});
OptionalIntegerFieldInput.displayName = "OptionalIntegerFieldInput";
