import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type CatalogPaginationProps = {
  className?: string;
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
};

export function CatalogPagination({
  className,
  page,
  totalPages,
  onPageChange,
}: CatalogPaginationProps) {
  if (totalPages <= 1) {
    return null;
  }

  return (
    <nav
      aria-label="Pagination"
      className={cn("catalog-pagination", className)}
    >
      <Button
        disabled={page <= 1}
        onClick={() => onPageChange(Math.max(1, page - 1))}
        size="sm"
        type="button"
        variant="outline"
      >
        Previous
      </Button>
      <p className="catalog-pagination__label text-caption">
        Page {page} of {totalPages}
      </p>
      <Button
        disabled={page >= totalPages}
        onClick={() => onPageChange(page + 1)}
        size="sm"
        type="button"
        variant="outline"
      >
        Next
      </Button>
    </nav>
  );
}
