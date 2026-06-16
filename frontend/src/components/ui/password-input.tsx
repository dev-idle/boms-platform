"use client";

import * as React from "react";
import { useId, useRef, useState } from "react";

import { useInputSelectionRestore } from "@/lib/hooks/use-input-selection-restore";
import { usePasswordMaskStrategy } from "@/lib/hooks/use-password-mask-strategy";
import { cn } from "@/lib/utils";

import { Input, type InputProps } from "./input";

export type PasswordInputProps = Omit<InputProps, "type">;

function EyeOpenIcon({ className }: { className?: string }) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={1.75}
      viewBox="0 0 24 24"
    >
      <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7Z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

function EyeClosedIcon({ className }: { className?: string }) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={1.75}
      viewBox="0 0 24 24"
    >
      <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
      <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68" />
      <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61" />
      <path d="m2 2 20 20" />
    </svg>
  );
}

function mergeRefs<T>(...refs: Array<React.Ref<T> | undefined>) {
  return (value: T | null) => {
    for (const ref of refs) {
      if (typeof ref === "function") {
        ref(value);
      } else if (ref) {
        (ref as React.MutableRefObject<T | null>).current = value;
      }
    }
  };
}

/**
 * Password field with a custom visibility toggle.
 *
 * Chromium / WebKit: always `type="text"` + CSS disc masking — no type flip, no caret jump.
 * Firefox: native `password` / `text` flip with selection restore.
 */
export const PasswordInput = React.forwardRef<HTMLInputElement, PasswordInputProps>(
  (
    {
      className,
      disabled,
      id,
      onChange,
      onClick,
      onInput,
      onKeyUp,
      onSelect,
      ...props
    },
    ref,
  ) => {
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
          {visible ? (
            <EyeClosedIcon className="password-field-toggle-icon" />
          ) : (
            <EyeOpenIcon className="password-field-toggle-icon" />
          )}
        </button>
      </div>
    );
  },
);
PasswordInput.displayName = "PasswordInput";
