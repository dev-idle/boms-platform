"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";

import { useCategories, useDeleteCategory } from "../hooks";

const PAGE_SIZE = 20;

export function ManagerCategoriesTable() {
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
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

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
            Categories
          </h1>
          <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">
            Organize the product catalog for customers.
          </p>
        </div>
        <Link href={ROUTE.manager.categoriesNew}>
          <Button type="button">New category</Button>
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
        />
        <Button type="submit">Search</Button>
      </form>

      <div className="overflow-x-auto rounded-lg border border-zinc-200 dark:border-zinc-800">
        <table className="min-w-full divide-y divide-zinc-200 text-sm dark:divide-zinc-800">
          <thead className="bg-zinc-50 dark:bg-zinc-900">
            <tr className="text-left text-xs uppercase tracking-wide text-zinc-500">
              <th className="px-4 py-3">Name</th>
              <th className="px-4 py-3">Slug</th>
              <th className="px-4 py-3">Order</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-200 dark:divide-zinc-800">
            {query.isPending ? (
              <tr>
                <td className="px-4 py-4 text-zinc-500" colSpan={5}>
                  Loading categories…
                </td>
              </tr>
            ) : categories.length === 0 ? (
              <tr>
                <td className="px-4 py-4 text-zinc-500" colSpan={5}>
                  No categories found.
                </td>
              </tr>
            ) : (
              categories.map((category) => (
                <tr key={category.id}>
                  <td className="px-4 py-3 font-medium">{category.name}</td>
                  <td className="px-4 py-3 text-zinc-600">{category.slug}</td>
                  <td className="px-4 py-3">{category.sort_order}</td>
                  <td className="px-4 py-3">
                    {category.is_active ? "Active" : "Inactive"}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <Link href={ROUTE.manager.categoryDetail(category.id)}>
                        <Button size="sm" type="button" variant="outline">
                          Edit
                        </Button>
                      </Link>
                      <Button
                        onClick={() =>
                          setDeleteTarget({
                            id: category.id,
                            name: category.name,
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
          <span className="text-sm text-zinc-600">
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
