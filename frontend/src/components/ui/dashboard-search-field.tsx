"use client";

import { useId } from "react";

import { DashboardCloseIcon } from "@/components/icons/dashboard-ui-icons";
import { cn } from "@/lib/utils";

type DashboardSearchFieldProps = {
  className?: string;
  id?: string;
  onChange: (value: string) => void;
  onClear: () => void;
  placeholder: string;
  value: string;
};

function SearchGlyph({ className }: { className?: string }) {
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
      <circle cx="11" cy="11" r="7" />
      <path d="M20 20 16.5 16.5" />
    </svg>
  );
}

/** Debounced-friendly search chrome for internal dashboard tables. */
export function DashboardSearchField({
  className,
  id,
  onChange,
  onClear,
  placeholder,
  value,
}: DashboardSearchFieldProps) {
  const fallbackId = useId();
  const inputId = id ?? fallbackId;

  return (
    <div className={cn("search-field", className)} role="search">
      <div className="search-field__box">
        <span aria-hidden className="search-field__leading">
          <SearchGlyph className="search-field__glyph" />
        </span>
        <input
          autoComplete="off"
          className="search-field__input"
          id={inputId}
          onChange={(event) => onChange(event.target.value)}
          placeholder={placeholder}
          spellCheck={false}
          type="search"
          value={value}
        />
        <button
          aria-label="Clear search"
          className={cn(
            "search-field__action",
            !value && "search-field__action--idle",
          )}
          onClick={onClear}
          tabIndex={value ? 0 : -1}
          type="button"
        >
          <DashboardCloseIcon className="size-3.5" />
        </button>
      </div>
    </div>
  );
}
