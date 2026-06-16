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
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import {
  DashboardTableDeleteButton,
  DashboardTableEditLink,
  DashboardTableRowActions,
} from "@/components/ui/dashboard-table-actions";
import { EntityActivePill } from "@/components/ui/status-pill";
import { DASHBOARD_TABLE_PAGE_SIZE } from "@/constants/dashboard-table";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { useDebouncedTableSearch } from "@/lib/hooks/use-debounced-table-search";
import { getDashboardQuerySurface } from "@/lib/react-query/query-surface";
import { paginatedPlaceholderCountFromMeta } from "@/lib/pagination/dashboard-pagination";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";

import { useCategories, useDeleteCategory } from "../hooks";

const PAGE_SIZE = DASHBOARD_TABLE_PAGE_SIZE;

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
  const { initialLoading, refetching } = getDashboardQuerySurface(query);
  const categories = query.data?.categories ?? [];
  const pagination = query.data?.pagination;
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
        <div className="db-table-filters">
          <DashboardSearchField
            onChange={setInput}
            onClear={clear}
            placeholder="Search name or slug"
            value={input}
          />
        </div>

        <DashboardTableWrap refetching={refetching}>
          <table className="db-table db-table--catalog db-table--categories db-table--comfortable">
            <colgroup>
              <col className="db-table-col-name" />
              <col className="db-table-col-slug" />
              <col className="db-table-col-order" />
              <col className="db-table-col-status" />
              <col className="db-table-col-actions" />
            </colgroup>
            <thead>
              <tr>
                <th>Name</th>
                <th>Slug</th>
                <th className="db-table-cell-order">Order</th>
                <th className="db-table-status">Status</th>
                <th className="db-table-detail">Actions</th>
              </tr>
            </thead>
            <tbody>
              <DashboardTableStateRows
                columnCount={5}
                entityLabel="categories"
                hasActiveFilter={Boolean(search)}
                isEmpty={categories.length === 0}
                isError={query.isError}
                initialLoading={initialLoading}
              />
              {!initialLoading && !query.isError && categories.length > 0
                ? categories.map((category) => (
                  <tr key={category.id}>
                    <td className="db-table-cell-primary">{category.name}</td>
                    <td
                      className="db-table-cell-truncate text-muted"
                      title={category.slug}
                    >
                      {category.slug}
                    </td>
                    <td className="db-table-cell-order">{category.sort_order}</td>
                    <td className="db-table-status">
                      <EntityActivePill active={category.is_active} />
                    </td>
                    <td className="db-table-detail">
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
                : null}
              <DashboardTablePagePlaceholders
                columnCount={5}
                count={pagePlaceholderCount}
              />
            </tbody>
          </table>
          <DashboardTablePagination
            disabled={query.isFetching}
            itemLabel="categories"
            onPageChange={setPage}
            page={pagination?.page ?? page}
            pageSize={pagination?.page_size ?? PAGE_SIZE}
            totalItems={pagination?.total ?? categories.length}
            totalPages={pagination?.total_pages ?? 1}
          />
        </DashboardTableWrap>
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
