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

import { useDeleteDiscountCode, useDiscountCodes } from "../hooks";
import { DISCOUNT_TYPE } from "../schemas";

const PAGE_SIZE = 20;

function formatDiscountValue(
  discountType: string,
  value: number,
): string {
  if (discountType === DISCOUNT_TYPE.percent) {
    return `${value}%`;
  }
  return formatPriceCents(value);
}

export function ManagerDiscountCodesTable() {
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string;
    code: string;
  } | null>(null);
  const deleteDiscountCode = useDeleteDiscountCode();

  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, search }),
    [page, search],
  );
  const query = useDiscountCodes(filter);
  const discountCodes = query.data?.discount_codes ?? [];
  const pagination = query.data?.pagination;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-page-title">
            Discount codes
          </h1>
          <p className="mt-2 text-sm text-ink-2">
            Create promotion codes for checkout (validated server-side at cart).
          </p>
        </div>
        <Link href={ROUTE.manager.discountCodesNew}>
          <Button type="button">New discount code</Button>
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
          placeholder="Search code"
          value={searchInput}
        />
        <Button type="submit">Search</Button>
      </form>

      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="min-w-full divide-y divide-border text-sm">
          <thead className="bg-surface-alt">
            <tr className="text-left text-xs uppercase tracking-wide text-muted">
              <th className="px-4 py-3">Code</th>
              <th className="px-4 py-3">Value</th>
              <th className="px-4 py-3">Uses</th>
              <th className="px-4 py-3">Window</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {query.isPending ? (
              <tr>
                <td className="px-4 py-6 text-muted" colSpan={6}>
                  Loading discount codes…
                </td>
              </tr>
            ) : query.isError ? (
              <tr>
                <td className="px-4 py-6 text-error" colSpan={6}>
                  Failed to load discount codes.
                </td>
              </tr>
            ) : discountCodes.length === 0 ? (
              <tr>
                <td className="px-4 py-6 text-muted" colSpan={6}>
                  No discount codes found.
                </td>
              </tr>
            ) : (
              discountCodes.map((discountCode) => (
                <tr key={discountCode.id}>
                  <td className="px-4 py-3 font-mono font-medium">
                    {discountCode.code}
                  </td>
                  <td className="px-4 py-3">
                    {formatDiscountValue(
                      discountCode.discount_type,
                      discountCode.value,
                    )}
                  </td>
                  <td className="px-4 py-3">
                    {discountCode.used_count}
                    {discountCode.max_uses != null
                      ? ` / ${discountCode.max_uses}`
                      : ""}
                  </td>
                  <td className="px-4 py-3 text-ink-2">
                    {formatDateTime(discountCode.starts_at)} –{" "}
                    {formatDateTime(discountCode.ends_at)}
                  </td>
                  <td className="px-4 py-3">
                    {discountCode.is_active ? "Active" : "Inactive"}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex flex-wrap gap-2">
                      <Link href={ROUTE.manager.discountCodeDetail(discountCode.id)}>
                        <Button size="sm" type="button" variant="outline">
                          Edit
                        </Button>
                      </Link>
                      <Button
                        onClick={() =>
                          setDeleteTarget({
                            id: discountCode.id,
                            code: discountCode.code,
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
        description={`This will remove "${deleteTarget?.code ?? "this discount code"}". This cannot be undone.`}
        isPending={deleteDiscountCode.isPending}
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => {
          if (!deleteTarget) {
            return;
          }
          deleteDiscountCode.mutate(deleteTarget.id, {
            onSuccess: () => setDeleteTarget(null),
            onError: (error) => {
              toast.error(
                isApiError(error)
                  ? error.message
                  : "Failed to delete discount code",
              );
            },
          });
        }}
        open={deleteTarget !== null}
        title="Delete discount code?"
      />
    </div>
  );
}
