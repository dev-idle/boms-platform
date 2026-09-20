"use client";

import * as React from "react";
import { useId, useRef, useState } from "react";

import { useInputSelectionRestore } from "@/lib/hooks/use-input-selection-restore";
import { usePasswordMaskStrategy } from "@/lib/hooks/use-password-mask-strategy";
import { cn } from "@/lib/utils";

import { Input, type InputProps } from "./input";

export type PasswordInputProps = Omit<InputProps, "type">;

function mergeRefs<T>(...refs: Array<React.Ref<T> | undefined>) {
  return (value: T | null) => {
    for (const ref of refs) {
      if (typeof ref === "function") {
        ref(value);
      } else if (ref) {
        (ref as React.RefObject<T | null>).current = value;
      }
    }
  };
}

/**
 * Password field with a text Show / Hide toggle — a word needs no legend.
 *
 * Chromium / WebKit: always `type="text"` + CSS disc masking — no type flip, no caret jump.
 * Firefox: native `password` / `text` flip with selection restore.
 */
export function PasswordInput({
  className,
  disabled,
  id,
  onChange,
  onClick,
  onInput,
  onKeyUp,
  onSelect,
  ref,
  ...props
}: PasswordInputProps) {
  const fallbackId = useId();
  const inputId = id ?? fallbackId;
  const inputRef = useRef<HTMLInputElement>(null);
  const maskStrategy = usePasswordMaskStrategy();
  const [visible, setVisible] = useState(false);

  const usesCssMask = maskStrategy === "css";
  const usesNativeTypeFlip = maskStrategy === "native";

  const { captureSelectionForToggle, rememberSelection } =
    useInputSelectionRestore({
      enabled: usesNativeTypeFlip,
      inputRef,
      restoreKey: visible,
    });

  function toggleVisibility(): void {
    setVisible((current) => !current);
  }

  const inputType = usesCssMask || visible ? "text" : "password";

  return (
    <div className="password-field">
      <Input
        ref={mergeRefs(ref, inputRef)}
        className={cn(
          "password-field-input",
          usesCssMask && !visible && "password-field-input--masked",
          className,
        )}
        disabled={disabled}
        id={inputId}
        spellCheck={false}
        type={inputType}
        onChange={(event) => {
          onChange?.(event);
          if (usesNativeTypeFlip) {
            rememberSelection(event.currentTarget);
          }
        }}
        onClick={(event) => {
          onClick?.(event);
          if (usesNativeTypeFlip) {
            rememberSelection(event.currentTarget);
          }
        }}
        onInput={(event) => {
          onInput?.(event);
          if (usesNativeTypeFlip) {
            rememberSelection(event.currentTarget);
          }
        }}
        onKeyUp={(event) => {
          onKeyUp?.(event);
          if (usesNativeTypeFlip) {
            rememberSelection(event.currentTarget);
          }
        }}
        onSelect={(event) => {
          onSelect?.(event);
          if (usesNativeTypeFlip) {
            rememberSelection(event.currentTarget);
          }
        }}
        {...props}
      />
      <button
        aria-controls={inputId}
        aria-label={visible ? "Hide password" : "Show password"}
        aria-pressed={visible}
        className="password-field-toggle"
        disabled={disabled}
        onClick={toggleVisibility}
        onPointerDown={(event) => {
          event.preventDefault();
          if (usesNativeTypeFlip) {
            captureSelectionForToggle();
          }
        }}
        type="button"
      >
        {visible ? "Hide" : "Show"}
      </button>
    </div>
  );
}
