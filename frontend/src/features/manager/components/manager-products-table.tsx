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
import { DashboardTableThumb } from "@/components/ui/dashboard-table-thumb";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import {
  DashboardTableDeleteButton,
  DashboardTableEditLink,
  DashboardTableRowActions,
} from "@/components/ui/dashboard-table-actions";
import { CatalogAvailabilityPill } from "@/components/ui/status-pill";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { DASHBOARD_TABLE_PAGE_SIZE } from "@/constants/dashboard-table";
import { DashboardTableToolbarMeta } from "@/components/ui/dashboard-table-toolbar-meta";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { useDebouncedTableSearch } from "@/lib/hooks/use-debounced-table-search";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { paginatedPlaceholderCountFromMeta } from "@/lib/pagination/dashboard-pagination";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
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
  const { initialLoading, refetching } = getQuerySurface(query);
  const products = query.data?.products ?? [];
  const pagination = query.data?.pagination;
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    products.length,
    pagination,
    PAGE_SIZE,
  );

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <Button asChild>
            <Link href={ROUTE.manager.productsNew}>+ New product</Link>
          </Button>
        }
        description="Manage items shown on the customer storefront."
        eyebrow={DASHBOARD_PAGE_EYEBROW.catalogue}
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
          {pagination ? (
            <DashboardTableToolbarMeta>
              {pagination.total}{" "}
              {pagination.total === 1 ? "product" : "products"}
            </DashboardTableToolbarMeta>
          ) : null}
        </div>

        <DashboardTableWrap refetching={refetching}>
          <table className="db-table db-table--catalog">
            <colgroup>
              <col className="db-table-col-thumb" />
            </colgroup>
            <thead>
              <tr>
                <th className="db-table-cell-thumb">
                  <span className="sr-only">Image</span>
                </th>
                <th>Product</th>
                <th>Category</th>
                <th className="db-table-num">Price</th>
                <th className="db-table-status">Status</th>
                <th className="db-table-detail">Actions</th>
              </tr>
            </thead>
            <tbody>
              <DashboardTableStateRows
                columnCount={6}
                entityLabel="products"
                hasActiveFilter={Boolean(search)}
                isEmpty={products.length === 0}
                isError={query.isError}
                initialLoading={initialLoading}
              />
              {!initialLoading && !query.isError && products.length > 0
                ? products.map((product) => (
                <tr key={product.id}>
                  <DashboardTableThumb
                    name={product.name}
                    url={product.image_urls[0]}
                  />
                  <td className="db-table-cell-primary">{product.name}</td>
                  <td className="text-muted">
                    {product.category_name ?? (
                      <span className="db-table-cell-placeholder">—</span>
                    )}
                  </td>
                  <td className="db-table-num">
                    {formatPriceCents(product.price_cents)}
                  </td>
                  <td className="db-table-status">
                    <CatalogAvailabilityPill available={product.is_active} />
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
                columnCount={6}
                count={pagePlaceholderCount}
              />
            </tbody>
          </table>
          <DashboardTablePagination
            disabled={refetching}
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
