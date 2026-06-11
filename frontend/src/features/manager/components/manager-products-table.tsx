"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useDeleteProduct, useProducts } from "../hooks";

const PAGE_SIZE = 20;

export function ManagerProductsTable() {
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
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

      <form
        className="flex flex-wrap items-center gap-2"
        onSubmit={(event) => {
          event.preventDefault();
          setPage(1);
          setSearch(searchInput.trim());
        }}
      >
        <Input
          className="max-w-sm"
          onChange={(event) => setSearchInput(event.target.value)}
          placeholder="Search name or slug"
          value={searchInput}
          variant="inline"
        />
        <Button type="submit">Search</Button>
      </form>

      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="min-w-full divide-y divide-border text-sm">
          <thead className="bg-surface-alt">
            <tr className="text-left text-xs uppercase tracking-wide text-muted">
              <th className="px-4 py-3">Name</th>
              <th className="px-4 py-3">Category</th>
              <th className="px-4 py-3">Price</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {query.isPending ? (
              <tr>
                <td className="px-4 py-4 text-muted" colSpan={5}>
                  Loading products…
                </td>
              </tr>
            ) : products.length === 0 ? (
              <tr>
                <td className="px-4 py-4 text-muted" colSpan={5}>
                  No products found.
                </td>
              </tr>
            ) : (
              products.map((product) => (
                <tr key={product.id}>
                  <td className="px-4 py-3 font-medium">{product.name}</td>
                  <td className="px-4 py-3 text-ink-2">
                    {product.category_name ?? "—"}
                  </td>
                  <td className="px-4 py-3">{formatPriceCents(product.price_cents)}</td>
                  <td className="px-4 py-3">
                    {product.is_available ? "Available" : "Unavailable"}
                  </td>
                  <td className="px-4 py-3">
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
      </div>

      {pagination && pagination.total_pages > 1 ? (
        <div className="flex items-center gap-2">
          <Button
            disabled={page <= 1}
            onClick={() => setPage((current) => Math.max(1, current - 1))}
            type="button"
            variant="outline"
          >
            Previous
          </Button>
          <span className="text-sm text-ink-2">
            Page {pagination.page} of {pagination.total_pages}
          </span>
          <Button
            disabled={page >= pagination.total_pages}
            onClick={() => setPage((current) => current + 1)}
            type="button"
            variant="outline"
          >
            Next
          </Button>
        </div>
      ) : null}

      <ConfirmDialog
        confirmLabel="Delete"
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
