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
import { CatalogAvailabilityPill } from "@/components/ui/status-pill";
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
import { formatPriceCents } from "@/lib/validation/catalog";

import { useDeleteProduct, useProducts } from "../hooks";

const PAGE_SIZE = 20;

export function ManagerProductsTable() {
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
  const deleteProduct = useDeleteProduct();

  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, search, category_id: "" }),
    [page, search],
  );
  const query = useProducts(filter);
  const products = query.data?.products ?? [];
  const pagination = query.data?.pagination;
  const initialLoad = isInitialQueryLoad(query.isPending, query.data);
  const refetching = isQueryRefetching(
    query.isFetching,
    query.isPending,
    query.data,
  );
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    products.length,
    pagination,
    PAGE_SIZE,
  );

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <Link href={ROUTE.manager.productsNew}>
            <Button type="button">New product</Button>
          </Link>
        }
        description="Manage items shown on the customer storefront."
        title={PAGE_TITLES.products}
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
              <th>Category</th>
              <th>Price</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {initialLoad ? (
              <tr>
                <td className="db-table-empty-cell" colSpan={5}>
                  Loading products…
                </td>
              </tr>
            ) : query.isError ? (
              <tr>
                <td className="db-table-empty-cell text-error" colSpan={5}>
                  Failed to load products.
                </td>
              </tr>
            ) : products.length === 0 ? (
              <tr>
                <td className="db-table-empty-cell" colSpan={5}>
                  No products found.
                </td>
              </tr>
            ) : (
              products.map((product) => (
                <tr key={product.id}>
                  <td className="db-table-cell-primary">{product.name}</td>
                  <td className="text-ink-2">
                    {product.category_name ?? (
                      <span className="db-table-cell-placeholder">—</span>
                    )}
                  </td>
                  <td className="text-tabular">
                    {formatPriceCents(product.price_cents)}
                  </td>
                  <td>
                    <CatalogAvailabilityPill available={product.is_available} />
                  </td>
                  <td>
                    <DashboardTableRowActions>
                      <DashboardTableEditLink
                        href={ROUTE.manager.productDetail(product.id)}
                        label={`Edit ${product.name}`}
                      />
                      <DashboardTableDeleteButton
                        label={`Delete ${product.name}`}
                        onClick={() =>
                          setDeleteTarget({
                            id: product.id,
                            name: product.name,
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
        description={`This will remove "${deleteTarget?.name ?? "this product"}". This cannot be undone.`}
        isPending={deleteProduct.isPending}
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => {
          if (!deleteTarget) {
            return;
          }
          deleteProduct.mutate(deleteTarget.id, {
            onSuccess: () => setDeleteTarget(null),
            onError: (error) => {
              toast.error(
                isApiError(error)
                  ? error.message
                  : "Failed to delete product",
              );
            },
          });
        }}
        open={deleteTarget !== null}
        title="Delete product?"
      />
    </div>
  );
}
