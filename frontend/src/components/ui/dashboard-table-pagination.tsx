"use client";

import { cn } from "@/lib/utils";

function ChevronIcon({ direction }: { direction: "left" | "right" }) {
  return (
    <svg
      aria-hidden
      className="dashboard-pagination-icon"
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={1.75}
      viewBox="0 0 24 24"
    >
      {direction === "left" ? (
        <path d="m15 18-6-6 6-6" />
      ) : (
        <path d="m9 18 6-6-6-6" />
      )}
    </svg>
  );
}

type DashboardTablePaginationProps = {
  className?: string;
  disabled?: boolean;
  hideWhenSinglePage?: boolean;
  onPageChange: (page: number) => void;
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
};

/** Prev/next control with page label for dashboard tables. */
export function DashboardTablePagination({
  className,
  disabled = false,
  hideWhenSinglePage = true,
  onPageChange,
  page,
  totalItems,
  totalPages,
}: DashboardTablePaginationProps) {
  const safeTotalPages = Math.max(totalPages, 1);
  const canPrevious = page > 1 && !disabled;
  const canNext = page < safeTotalPages && !disabled;

  if (hideWhenSinglePage && safeTotalPages <= 1 && totalItems > 0) {
    return null;
  }

  return (
    <nav
      aria-label="Table pagination"
      className={cn("dashboard-pagination", className)}
    >
      <div className="dashboard-pagination-controls" role="group" aria-label="Page navigation">
        <button
          aria-label="Previous page"
          className="dashboard-pagination-step dashboard-pagination-step--prev"
          disabled={!canPrevious}
          onClick={() => onPageChange(page - 1)}
          type="button"
        >
          <ChevronIcon direction="left" />
          <span>Prev</span>
        </button>
        <p aria-live="polite" className="dashboard-pagination-meta">
          Page {page} of {safeTotalPages}
        </p>
        <button
          aria-label="Next page"
          className="dashboard-pagination-step dashboard-pagination-step--next"
          disabled={!canNext}
          onClick={() => onPageChange(page + 1)}
          type="button"
        >
          <span>Next</span>
          <ChevronIcon direction="right" />
        </button>
      </div>
    </nav>
  );
}
