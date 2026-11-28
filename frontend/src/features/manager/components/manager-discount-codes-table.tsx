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
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { DASHBOARD_TABLE_PAGE_SIZE } from "@/constants/dashboard-table";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { useDebouncedTableSearch } from "@/lib/hooks/use-debounced-table-search";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { paginatedPlaceholderCountFromMeta } from "@/lib/pagination/dashboard-pagination";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
import { DashboardTableDateRangeCell } from "@/components/ui/dashboard-table-date-range-cell";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { keepPageInRange } from "@/lib/pagination/dashboard-pagination";
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
  const { initialLoading, refetching } = getQuerySurface(query);
  const discountCodes = query.data?.discount_codes ?? [];
  const pagination = query.data?.pagination;
  keepPageInRange(page, setPage, pagination?.total_pages);
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    discountCodes.length,
    pagination,
    PAGE_SIZE,
  );

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <Button asChild>
            <Link href={ROUTE.manager.discountCodesNew}>+ New discount code</Link>
          </Button>
        }
        description="Codes a customer enters at checkout for money or a percent off."
        eyebrow={DASHBOARD_PAGE_EYEBROW.promotions}
        title={PAGE_TITLES.discountCodes}
      />

      <div className="dashboard-page-body">
        <div className="db-table-filters">
          <DashboardSearchField
            label="Search discount codes"
            onChange={setInput}
            onClear={clear}
            placeholder="Search code"
            value={input}
          />
        </div>

        <DashboardTableWrap refetching={refetching}>
          <table className="db-table db-table--catalog db-table--discount-codes">
            <colgroup>
              <col className="db-table-col-code" />
              <col className="db-table-col-number" />
              <col className="db-table-col-number" />
              <col className="db-table-col-datetime" />
            </colgroup>
            <thead>
              <tr>
                <th scope="col">Code</th>
                <th scope="col" className="db-table-num">Value</th>
                <th scope="col" className="db-table-num">Uses</th>
                <th scope="col">Runs</th>
                <th scope="col" className="db-table-status">Status</th>
                <th scope="col" className="db-table-detail">Actions</th>
              </tr>
            </thead>
            <tbody>
              <DashboardTableStateRows
                columnCount={6}
                entityLabel="discount codes"
                hasActiveFilter={Boolean(search)}
                isEmpty={discountCodes.length === 0}
                isError={query.isError}
                onRetry={() => void query.refetch()}
                initialLoading={initialLoading}
              />
              {!initialLoading && !query.isError && discountCodes.length > 0
                ? discountCodes.map((discountCode) => (
                <tr key={discountCode.id}>
                  <td className="db-table-cell-primary text-order-code">
                    {discountCode.code}
                  </td>
                  <td className="db-table-num">
                    {formatDiscountValue(
                      discountCode.discount_type,
                      discountCode.value,
                    )}
                  </td>
                  <td className="db-table-num">
                    {discountCode.used_count}
                    {discountCode.max_uses != null
                      ? ` / ${discountCode.max_uses}`
                      : ""}
                  </td>
                  <DashboardTableDateRangeCell endsAt={discountCode.ends_at} startsAt={discountCode.starts_at} />
                  <td className="db-table-status">
                    <EntityActivePill active={discountCode.is_active} />
                  </td>
                  <td className="db-table-detail">
                    <DashboardTableRowActions>
                      <DashboardTableDeleteButton
                        label={`Delete ${discountCode.code}`}
                        onClick={() =>
                          setDeleteTarget({
                            id: discountCode.id,
                            code: discountCode.code,
                          })
                        }
                      />
                      <DashboardTableEditLink
                        href={ROUTE.manager.discountCodeDetail(discountCode.id)}
                        label={`Edit ${discountCode.code}`}
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
            itemLabel="discount codes"
            onPageChange={setPage}
            page={pagination?.page ?? page}
            pageSize={pagination?.page_size ?? PAGE_SIZE}
            totalItems={pagination?.total ?? discountCodes.length}
            totalPages={pagination?.total_pages ?? 1}
          />
        </DashboardTableWrap>
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
                  : "We could not delete discount code. Please try again.",
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
