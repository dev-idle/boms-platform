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
import { cn } from "@/lib/utils";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useDeleteDiscountCode, useDiscountCodes } from "../hooks";
import { DISCOUNT_TYPE } from "../schemas";

const PAGE_SIZE = DASHBOARD_TABLE_PAGE_SIZE;

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
  const initialLoad = isInitialQueryLoad(query.isPending, query.data);
  const refetching = isQueryRefetching(
    query.isFetching,
    query.isPending,
    query.data,
  );
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    discountCodes.length,
    pagination,
    PAGE_SIZE,
  );

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <Link href={ROUTE.manager.discountCodesNew}>
            <Button type="button">New discount code</Button>
          </Link>
        }
        description="Create promotion codes for checkout (validated server-side at cart)."
        title={PAGE_TITLES.discountCodes}
      />

      <div className="dashboard-page-body">
        <div className="db-table-filters">
          <DashboardSearchField
            onChange={setInput}
            onClear={clear}
            placeholder="Search code"
            value={input}
          />
        </div>

        <div className={cn("db-table-wrap", refetching && "is-refetching")}>
        <table className="db-table db-table--comfortable">
          <thead>
            <tr>
              <th>Code</th>
              <th>Value</th>
              <th>Uses</th>
              <th>Window</th>
              <th className="db-table-status">Status</th>
              <th className="db-table-detail">Actions</th>
            </tr>
          </thead>
          <tbody>
            {initialLoad ? (
              <tr>
                <td className="db-table-empty-cell" colSpan={6}>
                  Loading discount codes…
                </td>
              </tr>
            ) : query.isError ? (
              <tr>
                <td className="db-table-empty-cell text-error" colSpan={6}>
                  Failed to load discount codes.
                </td>
              </tr>
            ) : discountCodes.length === 0 ? (
              <tr>
                <td className="db-table-empty-cell" colSpan={6}>
                  {search
                    ? "No discount codes match your search."
                    : "No discount codes found."}
                </td>
              </tr>
            ) : (
              discountCodes.map((discountCode) => (
                <tr key={discountCode.id}>
                  <td className="db-table-cell-primary text-order-code">
                    {discountCode.code}
                  </td>
                  <td className="text-tabular">
                    {formatDiscountValue(
                      discountCode.discount_type,
                      discountCode.value,
                    )}
                  </td>
                  <td className="text-tabular">
                    {discountCode.used_count}
                    {discountCode.max_uses != null
                      ? ` / ${discountCode.max_uses}`
                      : ""}
                  </td>
                  <td className="text-muted">
                    {formatDateTime(discountCode.starts_at)} –{" "}
                    {formatDateTime(discountCode.ends_at)}
                  </td>
                  <td className="db-table-status">
                    <EntityActivePill active={discountCode.is_active} />
                  </td>
                  <td className="db-table-detail">
                    <DashboardTableRowActions>
                      <DashboardTableEditLink
                        href={ROUTE.manager.discountCodeDetail(discountCode.id)}
                        label={`Edit ${discountCode.code}`}
                      />
                      <DashboardTableDeleteButton
                        label={`Delete ${discountCode.code}`}
                        onClick={() =>
                          setDeleteTarget({
                            id: discountCode.id,
                            code: discountCode.code,
                          })
                        }
                      />
                    </DashboardTableRowActions>
                  </td>
                </tr>
              ))
            )}
            <DashboardTablePagePlaceholders
              columnCount={6}
              count={pagePlaceholderCount}
            />
          </tbody>
        </table>
        <DashboardTablePagination
          disabled={query.isFetching}
          itemLabel="discount codes"
          onPageChange={setPage}
          page={pagination?.page ?? page}
          pageSize={pagination?.page_size ?? PAGE_SIZE}
          totalItems={pagination?.total ?? discountCodes.length}
          totalPages={pagination?.total_pages ?? 1}
        />
      </div>
      </div>

      <ConfirmDialog
        confirmLabel="Delete"
        confirmVariant="destructive"
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
