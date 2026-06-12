"use client";

import * as React from "react";
import { useId, useState } from "react";

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

/** Password field with a single custom visibility toggle (Edge ::-ms-reveal suppressed). */
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

export const PasswordInput = React.forwardRef<HTMLInputElement, PasswordInputProps>(
  ({ className, disabled, id, ...props }, ref) => {
    const fallbackId = useId();
    const inputId = id ?? fallbackId;
    const inputRef = React.useRef<HTMLInputElement>(null);
    const [visible, setVisible] = useState(false);

    function toggleVisibility(): void {
      setVisible((current) => !current);
      inputRef.current?.focus();
    }

    return (
      <div className="password-field">
        <Input
          ref={mergeRefs(ref, inputRef)}
          className={cn("password-field-input", className)}
          disabled={disabled}
          id={inputId}
          spellCheck={false}
          type={visible ? "text" : "password"}
          {...props}
        />
        <button
          aria-controls={inputId}
          aria-label={visible ? "Hide password" : "Show password"}
          aria-pressed={visible}
          className="password-field-toggle"
          disabled={disabled}
          onClick={toggleVisibility}
          onMouseDown={(event) => {
            event.preventDefault();
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
