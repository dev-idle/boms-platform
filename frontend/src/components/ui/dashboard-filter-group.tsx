"use client";

import { cn } from "@/lib/utils";

export type DashboardFilterOption<T extends string | undefined> = {
  label: string;
  value: T;
};

type DashboardFilterGroupProps<T extends string | undefined> = {
  "aria-label": string;
  className?: string;
  onChange: (value: T) => void;
  options: ReadonlyArray<DashboardFilterOption<T>>;
  value: T;
};

/** Static segmented filters for dashboard tables (no motion). */
export function DashboardFilterGroup<T extends string | undefined>({
  "aria-label": ariaLabel,
  className,
  onChange,
  options,
  value,
}: DashboardFilterGroupProps<T>) {
  return (
    <div
      aria-label={ariaLabel}
      className={cn("db-filter-group", className)}
      role="group"
    >
      {options.map((option) => {
        const isActive = option.value === value;

        return (
          <button
            key={option.label}
            aria-pressed={isActive}
            className="db-filter-chip"
            data-active={isActive ? "true" : undefined}
            onClick={() => {
              onChange(option.value);
            }}
            type="button"
          >
            {option.label}
          </button>
        );
      })}
    </div>
  );
}
