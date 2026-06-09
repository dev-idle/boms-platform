import { Button } from "@/components/ui/button";

type CatalogPaginationProps = {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
};

export function CatalogPagination({
  page,
  totalPages,
  onPageChange,
}: CatalogPaginationProps) {
  if (totalPages <= 1) {
    return null;
  }

  return (
    <div className="flex items-center gap-3">
      <Button
        disabled={page <= 1}
        onClick={() => onPageChange(Math.max(1, page - 1))}
        type="button"
        variant="outline"
      >
        Previous
      </Button>
      <p className="text-sm text-muted">
        Page {page} of {totalPages}
      </p>
      <Button
        disabled={page >= totalPages}
        onClick={() => onPageChange(page + 1)}
        type="button"
        variant="outline"
      >
        Next
      </Button>
    </div>
  );
}
