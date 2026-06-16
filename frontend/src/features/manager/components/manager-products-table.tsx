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
import { CatalogAvailabilityPill } from "@/components/ui/status-pill";
import { DASHBOARD_TABLE_PAGE_SIZE } from "@/constants/dashboard-table";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { useDebouncedTableSearch } from "@/lib/hooks/use-debounced-table-search";
import { paginatedPlaceholderCountFromMeta } from "@/lib/pagination/dashboard-pagination";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
import {
  isInitialQueryLoad,
  isQueryRefetching,
} from "@/lib/react-query/query-surface";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useDeleteProduct, useProducts } from "../hooks";

const PAGE_SIZE = DASHBOARD_TABLE_PAGE_SIZE;

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
        <div className="db-table-filters">
          <DashboardSearchField
            onChange={setInput}
            onClear={clear}
            placeholder="Search name or slug"
            value={input}
          />
        </div>

        <DashboardTableWrap refetching={refetching}>
          <table className="db-table db-table--catalog db-table--comfortable">
            <thead>
              <tr>
                <th>Name</th>
                <th>Category</th>
                <th>Price</th>
                <th className="db-table-status">Status</th>
                <th className="db-table-detail">Actions</th>
              </tr>
            </thead>
            <tbody>
              <DashboardTableStateRows
                columnCount={5}
                entityLabel="products"
                hasActiveFilter={Boolean(search)}
                isEmpty={products.length === 0}
                isError={query.isError}
                isInitialLoad={initialLoad}
              />
              {!initialLoad && !query.isError && products.length > 0
                ? products.map((product) => (
                <tr key={product.id}>
                  <td className="db-table-cell-primary">{product.name}</td>
                  <td className="text-muted">
                    {product.category_name ?? (
                      <span className="db-table-cell-placeholder">—</span>
                    )}
                  </td>
                  <td className="text-tabular">
                    {formatPriceCents(product.price_cents)}
                  </td>
                  <td className="db-table-status">
                    <CatalogAvailabilityPill available={product.is_available} />
                  </td>
                  <td className="db-table-detail">
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
                : null}
              <DashboardTablePagePlaceholders
                columnCount={5}
                count={pagePlaceholderCount}
              />
            </tbody>
          </table>
          <DashboardTablePagination
            disabled={query.isFetching}
            itemLabel="products"
            onPageChange={setPage}
            page={pagination?.page ?? page}
            pageSize={pagination?.page_size ?? PAGE_SIZE}
            totalItems={pagination?.total ?? products.length}
            totalPages={pagination?.total_pages ?? 1}
          />
        </DashboardTableWrap>
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
