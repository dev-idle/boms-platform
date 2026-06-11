"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { DashboardSearchField } from "@/components/ui/dashboard-search-field";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { useDebouncedTableSearch } from "@/lib/hooks/use-debounced-table-search";
import {
  isInitialQueryLoad,
  isQueryRefetching,
} from "@/lib/react-query/query-surface";
import { cn } from "@/lib/utils";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useCombos, useDeleteCombo } from "../hooks";

const PAGE_SIZE = 20;

export function ManagerCombosTable() {
  const {
    clear,
    input,
    page,
    search,
    setInput,
    setPage,
  } = useDebouncedTableSearch();
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string;
    name: string;
  } | null>(null);
  const deleteCombo = useDeleteCombo();

  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, search }),
    [page, search],
  );
  const query = useCombos(filter);
  const combos = query.data?.combos ?? [];
  const pagination = query.data?.pagination;
  const initialLoad = isInitialQueryLoad(query.isPending, query.data);
  const refetching = isQueryRefetching(
    query.isFetching,
    query.isPending,
    query.data,
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-page-title">
            Combos
          </h1>
          <p className="mt-2 text-sm text-ink-2">
            Bundle products with promotional pricing and time windows.
          </p>
        </div>
        <Link href={ROUTE.manager.combosNew}>
          <Button type="button">New combo</Button>
        </Link>
      </div>

      <DashboardSearchField
        onChange={setInput}
        onClear={clear}
        placeholder="Search name or slug"
        value={input}
      />

      <div className={cn("db-table-wrap", refetching && "is-refetching")}>
        <table className="db-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Price</th>
              <th>Window</th>
              <th>Items</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {initialLoad ? (
              <tr>
                <td className="text-muted" colSpan={6}>
                  Loading combos…
                </td>
              </tr>
            ) : query.isError ? (
              <tr>
                <td className="text-error" colSpan={6}>
                  Failed to load combos.
                </td>
              </tr>
            ) : combos.length === 0 ? (
              <tr>
                <td className="text-muted" colSpan={6}>
                  No combos found.
                </td>
              </tr>
            ) : (
              combos.map((combo) => (
                <tr key={combo.id}>
                  <td className="font-medium">{combo.name}</td>
                  <td className="text-tabular">
                    {formatPriceCents(combo.price_cents)}
                  </td>
                  <td className="text-ink-2">
                    {formatDateTime(combo.starts_at)} – {formatDateTime(combo.ends_at)}
                  </td>
                  <td className="text-tabular">{combo.items.length}</td>
                  <td>{combo.is_active ? "Active" : "Inactive"}</td>
                  <td>
                    <div className="flex flex-wrap gap-2">
                      <Link href={ROUTE.manager.comboDetail(combo.id)}>
                        <Button size="sm" type="button" variant="outline">
                          Edit
                        </Button>
                      </Link>
                      <Button
                        onClick={() =>
                          setDeleteTarget({ id: combo.id, name: combo.name })
                        }
                        size="sm"
                        type="button"
                        variant="outline"
                      >
                        Delete
                      </Button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
        {pagination ? (
          <DashboardTablePagination
            disabled={query.isFetching}
            onPageChange={setPage}
            page={pagination.page}
            pageSize={pagination.page_size}
            totalItems={pagination.total}
            totalPages={pagination.total_pages}
          />
        ) : null}
      </div>

      <ConfirmDialog
        confirmLabel="Delete"
        confirmVariant="destructive"
        description={`This will remove "${deleteTarget?.name ?? "this combo"}". This cannot be undone.`}
        isPending={deleteCombo.isPending}
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => {
          if (!deleteTarget) {
            return;
          }
          deleteCombo.mutate(deleteTarget.id, {
            onSuccess: () => setDeleteTarget(null),
            onError: (error) => {
              toast.error(
                isApiError(error)
                  ? error.message
                  : "Failed to delete combo",
              );
            },
          });
        }}
        open={deleteTarget !== null}
        title="Delete combo?"
      />
    </div>
  );
}
