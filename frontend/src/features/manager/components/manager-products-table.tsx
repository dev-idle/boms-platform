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

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-page-title">
            Products
          </h1>
          <p className="mt-2 text-sm text-ink-2">
            Manage items shown on the customer storefront.
          </p>
        </div>
        <Link href={ROUTE.manager.productsNew}>
          <Button type="button">New product</Button>
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
              <th>Category</th>
              <th>Price</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {initialLoad ? (
              <tr>
                <td className="text-muted" colSpan={5}>
                  Loading products…
                </td>
              </tr>
            ) : products.length === 0 ? (
              <tr>
                <td className="text-muted" colSpan={5}>
                  No products found.
                </td>
              </tr>
            ) : (
              products.map((product) => (
                <tr key={product.id}>
                  <td className="font-medium">{product.name}</td>
                  <td className="text-ink-2">{product.category_name ?? "—"}</td>
                  <td className="text-tabular">
                    {formatPriceCents(product.price_cents)}
                  </td>
                  <td>{product.is_available ? "Available" : "Unavailable"}</td>
                  <td>
                    <div className="flex gap-2">
                      <Link href={ROUTE.manager.productDetail(product.id)}>
                        <Button size="sm" type="button" variant="outline">
                          Edit
                        </Button>
                      </Link>
                      <Button
                        onClick={() =>
                          setDeleteTarget({
                            id: product.id,
                            name: product.name,
                          })
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
