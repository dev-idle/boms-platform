"use client";

import type { ReactNode } from "react";

import {
  buildPaginationWindow,
  formatPaginationRange,
  paginationRange,
} from "@/lib/pagination/dashboard-pagination";
import { cn } from "@/lib/utils";

function PaginationStep({
  children,
  className,
  disabled = false,
  label,
  onClick,
}: {
  children: ReactNode;
  className?: string;
  disabled?: boolean;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      aria-label={label}
      className={cn("dashboard-pagination-step", className)}
      disabled={disabled}
      onClick={onClick}
      type="button"
    >
      {children}
    </button>
  );
}

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
  itemLabel?: string;
  onPageChange: (page: number) => void;
  page: number;
  pageSize: number;
  showPageNumbers?: boolean;
  showRange?: boolean;
  totalItems: number;
  totalPages: number;
};

/** Dashboard table footer — range summary, page window, prev/next. */
export function DashboardTablePagination({
  className,
  disabled = false,
  hideWhenSinglePage = true,
  itemLabel = "items",
  onPageChange,
  page,
  pageSize,
  showPageNumbers = true,
  showRange = true,
  totalItems,
  totalPages,
}: DashboardTablePaginationProps) {
  const safeTotalPages = Math.max(totalPages, 1);
  const canPrevious = page > 1 && !disabled;
  const canNext = page < safeTotalPages && !disabled;
  const range = paginationRange(page, pageSize, totalItems);
  const rangeLabel = formatPaginationRange(range, totalItems, itemLabel);
  const pageWindow = buildPaginationWindow(page, safeTotalPages);

  if (hideWhenSinglePage && (totalItems === 0 || safeTotalPages <= 1)) {
    return null;
  }

  return (
    <nav
      aria-label="Table pagination"
      className={cn("dashboard-pagination", className)}
    >
      {showRange ? (
        <p className="dashboard-pagination-summary">{rangeLabel}</p>
      ) : (
        <span aria-hidden className="dashboard-pagination-summary-spacer" />
      )}

      <div className="dashboard-pagination-controls" role="group" aria-label="Page navigation">
        <PaginationStep
          className="dashboard-pagination-step--prev"
          disabled={!canPrevious}
          label="Previous page"
          onClick={() => onPageChange(page - 1)}
        >
          <ChevronIcon direction="left" />
          <span>Prev</span>
        </PaginationStep>

        {showPageNumbers && safeTotalPages > 1 ? (
          <div aria-label="Page numbers" className="dashboard-pagination-pages">
            {pageWindow.map((item, index) =>
              item === "ellipsis" ? (
                <span
                  key={`ellipsis-${index}`}
                  aria-hidden
                  className="dashboard-pagination-ellipsis"
                >
                  …
                </span>
              ) : (
                <button
                  key={item}
                  aria-current={item === page ? "page" : undefined}
                  aria-label={`Page ${item}`}
                  className={cn(
                    "dashboard-pagination-page",
                    item === page && "dashboard-pagination-page--active",
                  )}
                  disabled={disabled || item === page}
                  onClick={() => onPageChange(item)}
                  type="button"
                >
                  {item}
                </button>
              ),
            )}
          </div>
        ) : (
          <p
            aria-live="polite"
            className="dashboard-pagination-meta dashboard-pagination-meta--slot"
          >
            Page {page} of {safeTotalPages}
          </p>
        )}

        <PaginationStep
          className="dashboard-pagination-step--next"
          disabled={!canNext}
          label="Next page"
          onClick={() => onPageChange(page + 1)}
        >
          <span>Next</span>
          <ChevronIcon direction="right" />
        </PaginationStep>
      </div>
    </nav>
  );
}
