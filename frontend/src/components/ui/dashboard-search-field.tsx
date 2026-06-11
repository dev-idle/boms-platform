"use client";

import { useId } from "react";

import { CloseIcon } from "@/components/icons/storefront-icons";
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
      strokeWidth={1.5}
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
    <div className={cn("dashboard-search", className)} role="search">
      <div className="dashboard-search-field">
        <span aria-hidden className="dashboard-search-leading">
          <SearchGlyph className="dashboard-search-glyph" />
        </span>
        <input
          autoComplete="off"
          className="dashboard-search-input"
          id={inputId}
          onChange={(event) => onChange(event.target.value)}
          placeholder={placeholder}
          spellCheck={false}
          type="search"
          value={value}
        />
        <button
          aria-label="Clear search"
          className={cn("dashboard-search-clear", !value && "is-empty")}
          onClick={onClear}
          tabIndex={value ? 0 : -1}
          type="button"
        >
          <CloseIcon className="size-3.5" />
        </button>
      </div>
    </div>
  );
}
