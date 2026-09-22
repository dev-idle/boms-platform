import type * as React from "react";

import {
  formatVietnamPhone,
  normalizeVietnamPhone,
} from "@/lib/validation/phone";

import { Input } from "./input";

/** Room for the longest spelling people type: `+84 0912 345 678` plus slack. */
const PHONE_INPUT_MAX_LENGTH = 20;

type PhoneInputProps = Omit<
  React.ComponentProps<"input">,
  "onChange" | "type" | "value"
> & {
  onChange: (value: string) => void;
  value: string | null | undefined;
};

/**
 * The one phone field. It takes a number however it is typed and, once the field
 * loses focus, tidies a valid one into the grouping people read (`0912 345 678`).
 * It never rewrites while the caret is inside — reformatting mid-keystroke moves
 * the caret — and leaves a value it cannot read untouched, so the validation
 * message still matches what is on screen. An empty field stays "": each schema
 * decides whether that means "clear the phone" or "no phone".
 */
export function PhoneInput({ onBlur, onChange, value, ...props }: PhoneInputProps) {
  return (
    <Input
      autoComplete="tel"
      inputMode="tel"
      maxLength={PHONE_INPUT_MAX_LENGTH}
      placeholder="0912 345 678"
      type="tel"
      {...props}
      onBlur={(event) => {
        const stored = normalizeVietnamPhone(event.target.value);
        if (stored) {
          onChange(formatVietnamPhone(stored));
        }
        onBlur?.(event);
      }}
      onChange={(event) => onChange(event.target.value)}
      value={value ?? ""}
    />
  );
}
