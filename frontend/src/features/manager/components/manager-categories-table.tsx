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

import { useCategories, useDeleteCategory } from "../hooks";

const PAGE_SIZE = 20;

export function ManagerCategoriesTable() {
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
  const deleteCategory = useDeleteCategory();

  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, search }),
    [page, search],
  );
  const query = useCategories(filter);
  const categories = query.data?.categories ?? [];
  const pagination = query.data?.pagination;
  const initialLoad = isInitialQueryLoad(query.isPending, query.data);
  const refetching = isQueryRefetching(
    query.isFetching,
    query.isPending,
    query.data,
  );
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    categories.length,
    pagination,
    PAGE_SIZE,
  );

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <Link href={ROUTE.manager.categoriesNew}>
            <Button type="button">New category</Button>
          </Link>
        }
        description="Organize the product catalog for customers."
        title={PAGE_TITLES.categories}
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
              <th>Slug</th>
              <th>Order</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {initialLoad ? (
              <tr>
                <td className="db-table-empty-cell" colSpan={5}>
                  Loading categories…
                </td>
              </tr>
            ) : query.isError ? (
              <tr>
                <td className="db-table-empty-cell text-error" colSpan={5}>
                  Failed to load categories.
                </td>
              </tr>
            ) : categories.length === 0 ? (
              <tr>
                <td className="db-table-empty-cell" colSpan={5}>
                  No categories found.
                </td>
              </tr>
            ) : (
              categories.map((category) => (
                <tr key={category.id}>
                  <td className="db-table-cell-primary">{category.name}</td>
                  <td className="text-ink-2">{category.slug}</td>
                  <td className="text-tabular">{category.sort_order}</td>
                  <td>
                    <EntityActivePill active={category.is_active} />
                  </td>
                  <td>
                    <DashboardTableRowActions>
                      <DashboardTableEditLink
                        href={ROUTE.manager.categoryDetail(category.id)}
                        label={`Edit ${category.name}`}
                      />
                      <DashboardTableDeleteButton
                        label={`Delete ${category.name}`}
                        onClick={() =>
                          setDeleteTarget({
                            id: category.id,
                            name: category.name,
                          })
                        }
                      />
                    </DashboardTableRowActions>
                  </td>
                </tr>
              ))
            )}
            <DashboardTablePagePlaceholders
              columnCount={5}
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
        description={`This will remove "${deleteTarget?.name ?? "this category"}". Products must be removed first.`}
        isPending={deleteCategory.isPending}
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => {
          if (!deleteTarget) {
            return;
          }
          deleteCategory.mutate(deleteTarget.id, {
            onSuccess: () => setDeleteTarget(null),
            onError: (error) => {
              if (
                isApiError(error) &&
                error.code === "category_has_products"
              ) {
                toast.error(
                  "Remove products before deleting this category.",
                );
                return;
              }
              toast.error(
                isApiError(error)
                  ? error.message
                  : "Failed to delete category",
              );
            },
          });
        }}
        open={deleteTarget !== null}
        title="Delete category?"
      />
    </div>
  );
}
