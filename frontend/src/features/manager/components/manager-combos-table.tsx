"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useCombos, useDeleteCombo } from "../hooks";

const PAGE_SIZE = 20;

export function ManagerCombosTable() {
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
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

      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="min-w-full divide-y divide-border text-sm">
          <thead className="bg-surface-alt">
            <tr className="text-left text-xs uppercase tracking-wide text-muted">
              <th className="px-4 py-3">Name</th>
              <th className="px-4 py-3">Price</th>
              <th className="px-4 py-3">Window</th>
              <th className="px-4 py-3">Items</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {query.isPending ? (
              <tr>
                <td className="px-4 py-6 text-muted" colSpan={6}>
                  Loading combos…
                </td>
              </tr>
            ) : query.isError ? (
              <tr>
                <td className="px-4 py-6 text-error" colSpan={6}>
                  Failed to load combos.
                </td>
              </tr>
            ) : combos.length === 0 ? (
              <tr>
                <td className="px-4 py-6 text-muted" colSpan={6}>
                  No combos found.
                </td>
              </tr>
            ) : (
              combos.map((combo) => (
                <tr key={combo.id}>
                  <td className="px-4 py-3 font-medium">{combo.name}</td>
                  <td className="px-4 py-3">{formatPriceCents(combo.price_cents)}</td>
                  <td className="px-4 py-3 text-ink-2">
                    {formatDateTime(combo.starts_at)} – {formatDateTime(combo.ends_at)}
                  </td>
                  <td className="px-4 py-3">{combo.items.length}</td>
                  <td className="px-4 py-3">
                    {combo.is_active ? "Active" : "Inactive"}
                  </td>
                  <td className="px-4 py-3">
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
