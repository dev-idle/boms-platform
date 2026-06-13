"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DashboardSearchField } from "@/components/ui/dashboard-search-field";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { DashboardTablePagePlaceholders } from "@/components/ui/dashboard-table-page-placeholders";
import {
  DashboardTableDeleteButton,
  DashboardTableEditLink,
  DashboardTableRowActions,
} from "@/components/ui/dashboard-table-actions";
import { EntityActivePill } from "@/components/ui/status-pill";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { useDebouncedTableSearch } from "@/lib/hooks/use-debounced-table-search";
import { paginatedPlaceholderCountFromMeta } from "@/lib/pagination/dashboard-pagination";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
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
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    combos.length,
    pagination,
    PAGE_SIZE,
  );

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <Link href={ROUTE.manager.combosNew}>
            <Button type="button">New combo</Button>
          </Link>
        }
        description="Bundle products with promotional pricing and time windows."
        title={PAGE_TITLES.combos}
      />

      <div className="dashboard-page-body">
      <DashboardSearchField
        onChange={setInput}
        onClear={clear}
        placeholder="Search name or slug"
        value={input}
      />

      <div className={cn("db-table-wrap", refetching && "is-refetching")}>
        <table className="db-table db-table--comfortable">
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
                <td className="db-table-empty-cell" colSpan={6}>
                  Loading combos…
                </td>
              </tr>
            ) : query.isError ? (
              <tr>
                <td className="db-table-empty-cell text-error" colSpan={6}>
                  Failed to load combos.
                </td>
              </tr>
            ) : combos.length === 0 ? (
              <tr>
                <td className="db-table-empty-cell" colSpan={6}>
                  No combos found.
                </td>
              </tr>
            ) : (
              combos.map((combo) => (
                <tr key={combo.id}>
                  <td className="db-table-cell-primary">{combo.name}</td>
                  <td className="text-tabular">
                    {formatPriceCents(combo.price_cents)}
                  </td>
                  <td className="text-ink-2">
                    {formatDateTime(combo.starts_at)} – {formatDateTime(combo.ends_at)}
                  </td>
                  <td className="text-tabular">{combo.items.length}</td>
                  <td>
                    <EntityActivePill active={combo.is_active} />
                  </td>
                  <td>
                    <DashboardTableRowActions>
                      <DashboardTableEditLink
                        href={ROUTE.manager.comboDetail(combo.id)}
                        label={`Edit ${combo.name}`}
                      />
                      <DashboardTableDeleteButton
                        label={`Delete ${combo.name}`}
                        onClick={() =>
                          setDeleteTarget({ id: combo.id, name: combo.name })
                        }
                      />
                    </DashboardTableRowActions>
                  </td>
                </tr>
              ))
            )}
            <DashboardTablePagePlaceholders
              columnCount={6}
              count={pagePlaceholderCount}
            />
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
