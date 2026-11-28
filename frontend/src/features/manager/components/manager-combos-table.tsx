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
import { DashboardTableThumb } from "@/components/ui/dashboard-table-thumb";
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

import { useCombos, useDeleteCombo } from "../hooks";

const PAGE_SIZE = DASHBOARD_TABLE_PAGE_SIZE;

export function ManagerCombosTable() {
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
  const deleteCombo = useDeleteCombo();

  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, search }),
    [page, search],
  );
  const query = useCombos(filter);
  const { initialLoading, refetching } = getQuerySurface(query);
  const combos = query.data?.combos ?? [];
  const pagination = query.data?.pagination;
  keepPageInRange(page, setPage, pagination?.total_pages);
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    combos.length,
    pagination,
    PAGE_SIZE,
  );

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <Button asChild>
            <Link href={ROUTE.manager.combosNew}>+ New combo</Link>
          </Button>
        }
        description="Bundle products with promotional pricing and time windows."
        eyebrow={DASHBOARD_PAGE_EYEBROW.promotions}
        title={PAGE_TITLES.combos}
      />

      <div className="dashboard-page-body">
        <div className="db-table-filters">
          <DashboardSearchField
            label="Search combos"
            onChange={setInput}
            onClear={clear}
            placeholder="Search name or slug"
            value={input}
          />
        </div>

        <DashboardTableWrap refetching={refetching}>
          <table className="db-table db-table--catalog db-table--combos">
            <colgroup>
              <col className="db-table-col-thumb" />
              <col />
              <col className="db-table-col-number" />
              <col className="db-table-col-count" />
              <col className="db-table-col-datetime" />
            </colgroup>
            <thead>
              <tr>
                <th scope="col" className="db-table-cell-thumb">
                  <span className="sr-only">Image</span>
                </th>
                <th scope="col">Name</th>
                <th scope="col" className="db-table-num">Price</th>
                <th scope="col" className="db-table-num">Items</th>
                <th scope="col">Runs</th>
                <th scope="col" className="db-table-status">Status</th>
                <th scope="col" className="db-table-detail">Actions</th>
              </tr>
            </thead>
            <tbody>
              <DashboardTableStateRows
                columnCount={7}
                entityLabel="combos"
                hasActiveFilter={Boolean(search)}
                isEmpty={combos.length === 0}
                isError={query.isError}
                onRetry={() => void query.refetch()}
                initialLoading={initialLoading}
              />
              {!initialLoading && !query.isError && combos.length > 0
                ? combos.map((combo) => (
                <tr key={combo.id}>
                  <DashboardTableThumb name={combo.name} url={combo.image_url} />
                  <td className="db-table-cell-primary">{combo.name}</td>
                  <td className="db-table-num">
                    {formatPriceCents(combo.price_cents)}
                  </td>
                  <td className="db-table-num">{combo.items.length}</td>
                  <DashboardTableDateRangeCell endsAt={combo.ends_at} startsAt={combo.starts_at} />
                  <td className="db-table-status">
                    <EntityActivePill active={combo.is_active} />
                  </td>
                  <td className="db-table-detail">
                    <DashboardTableRowActions>
                      <DashboardTableDeleteButton
                        label={`Delete ${combo.name}`}
                        onClick={() =>
                          setDeleteTarget({ id: combo.id, name: combo.name })
                        }
                      />
                      <DashboardTableEditLink
                        href={ROUTE.manager.comboDetail(combo.id)}
                        label={`Edit ${combo.name}`}
                      />
                    </DashboardTableRowActions>
                  </td>
                </tr>
              ))
                : null}
              <DashboardTablePagePlaceholders
                columnCount={7}
                count={pagePlaceholderCount}
              />
            </tbody>
          </table>
          <DashboardTablePagination
            disabled={refetching}
            itemLabel="combos"
            onPageChange={setPage}
            page={pagination?.page ?? page}
            pageSize={pagination?.page_size ?? PAGE_SIZE}
            totalItems={pagination?.total ?? combos.length}
            totalPages={pagination?.total_pages ?? 1}
          />
        </DashboardTableWrap>
      </div>

      <ConfirmDialog
        confirmLabel="Delete"
        confirmVariant="destructive"
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
                  : "We could not delete combo. Please try again.",
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
