"use client";

import * as React from "react";

import {
  acceptDollarDraft,
  formatDollarInput,
  parseDollarInput,
} from "@/lib/validation/catalog";
import { cn } from "@/lib/utils";

import { Input, type InputProps } from "./input";

type MoneyInputProps = Omit<
  InputProps,
  "type" | "value" | "onChange" | "onBlur" | "inputMode"
> & {
  onBlur?: () => void;
};

type DollarInputProps = MoneyInputProps & {
  /** The stored amount; null or undefined when the field holds none. */
  cents: number | null | undefined;
  commit: (raw: string) => void;
};

/**
 * A dollar amount, held as the integer cents the API stores.
 *
 * While the field has focus it shows exactly what was typed — normalising
 * mid-keystroke would fight the person entering "12.5". On blur it settles to
 * two decimals, so what is saved is what is read back.
 *
 * A keystroke that would not leave an amount is refused rather than parsed and
 * dropped: the field can never show `0.00a` while the form holds nothing. And
 * focusing selects the amount, so typing over a price replaces it instead of
 * appending to it.
 */
function DollarInput({
  className,
  cents,
  commit,
  onBlur,
  onFocus,
  placeholder = "0.00",
  ref,
  ...props
}: DollarInputProps) {
  const [draft, setDraft] = React.useState<string | null>(null);
  const selectingOnFocus = React.useRef(false);
  const value = draft ?? (cents == null ? "" : formatDollarInput(cents));

  return (
    <span className="money-input">
      <span aria-hidden className="money-input__prefix">
        $
      </span>
      <Input
        {...props}
        className={cn("money-input__control", className)}
        inputMode="decimal"
        onBlur={(event) => {
          commit(event.currentTarget.value);
          setDraft(null);
          // A field left by keyboard never sees the mouseup that would clear this,
          // and a stale flag would swallow the caret of the next click.
          selectingOnFocus.current = false;
          onBlur?.();
        }}
        onChange={(event) => {
          const next = acceptDollarDraft(event.target.value);
          if (next === null) {
            // Refused, so the draft does not change and React has nothing to
            // re-render — put the field back to the text it had itself.
            event.target.value = value;
            return;
          }
          setDraft(next);
          commit(next);
        }}
        onFocus={(event) => {
          setDraft(value);
          event.currentTarget.select();
          selectingOnFocus.current = true;
          onFocus?.(event);
        }}
        onMouseUp={(event) => {
          // The click that focused the field ends by dropping that selection.
          if (selectingOnFocus.current) {
            selectingOnFocus.current = false;
            event.preventDefault();
          }
          props.onMouseUp?.(event);
        }}
        placeholder={placeholder}
        ref={ref}
        value={value}
      />
    </span>
  );
}

/** Required amount — an empty or unreadable field reports no value, so the form validates. */
export function MoneyFieldInput({
  onChange,
  value,
  ...props
}: MoneyInputProps & {
  onChange: (cents: number | undefined) => void;
  value: number | undefined;
}) {
  return (
    <DollarInput
      {...props}
      cents={value}
      commit={(raw) => onChange(parseDollarInput(raw) ?? undefined)}
    />
  );
}

/** Optional amount — an empty field means "no limit", as the hints say. */
export function OptionalMoneyFieldInput({
  onChange,
  value,
  ...props
}: MoneyInputProps & {
  onChange: (cents: number | null) => void;
  value: number | null;
}) {
  return (
    <DollarInput
      {...props}
      cents={value}
      commit={(raw) =>
        onChange(raw.trim() === "" ? null : (parseDollarInput(raw) ?? null))
      }
    />
  );
}
