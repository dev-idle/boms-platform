"use client";

import Link from "next/link";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DashboardTableDateTimeCell } from "@/components/ui/dashboard-table-datetime-cell";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { StatusPill } from "@/components/ui/status-pill";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { DASHBOARD_TABLE_PAGE_SIZE } from "@/constants/dashboard-table";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
import { getQuerySurface } from "@/lib/react-query/query-surface";

import { usePromotions } from "../hooks";

const PAGE_SIZE = DASHBOARD_TABLE_PAGE_SIZE;
const COLUMN_COUNT = 5;

/** The promotions managers sent, latest first: who sent each, when, and to how many. */
export function ManagerPromotionsTable() {
  const [page, setPage] = useState(1);
  const promotionsQuery = usePromotions({ page, page_size: PAGE_SIZE });
  const { initialLoading, refetching } = getQuerySurface(promotionsQuery);
  const promotions = promotionsQuery.data?.promotions ?? [];
  const pagination = promotionsQuery.data?.pagination;

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <Button asChild>
            <Link href={ROUTE.manager.promotionsNew}>+ New promotion</Link>
          </Button>
        }
        description="Email offers to the customers who asked for them. Each promotion goes to everyone who has agreed and confirmed their address when it is sent."
        eyebrow={DASHBOARD_PAGE_EYEBROW.promotions}
        title={PAGE_TITLES.promotions}
      />

      <div className="dashboard-page-body">
        <DashboardTableWrap refetching={refetching}>
          <table className="db-table db-table--promotions">
            <colgroup>
              <col />
              <col className="db-table-col-sender" />
              <col className="db-table-col-datetime" />
              <col className="db-table-col-number" />
              <col className="db-table-col-status" />
            </colgroup>
            <thead>
              <tr>
                <th>Subject</th>
                <th>Sent by</th>
                <th>Sent</th>
                <th className="db-table-num">Recipients</th>
                <th className="db-table-status">Status</th>
              </tr>
            </thead>
            <tbody>
              <DashboardTableStateRows
                columnCount={COLUMN_COUNT}
                emptyMessage="No promotion sent yet."
                entityLabel="promotions"
                isEmpty={promotions.length === 0}
                isError={promotionsQuery.isError}
                initialLoading={initialLoading}
              />
              {!initialLoading && !promotionsQuery.isError
                ? promotions.map((promotion) => (
                    <tr key={promotion.id}>
                      <td className="db-table-cell-primary">{promotion.subject}</td>
                      <td className="text-muted">{promotion.sender_name ?? "—"}</td>
                      <DashboardTableDateTimeCell iso={promotion.created_at} />
                      <td className="db-table-num">{promotion.recipient_count ?? "—"}</td>
                      <td className="db-table-status">
                        <StatusPill
                          label={promotion.status === "sent" ? "Sent" : "Sending"}
                          variant={promotion.status === "sent" ? "completed" : "in_progress"}
                        />
                      </td>
                    </tr>
                  ))
                : null}
            </tbody>
          </table>
          <DashboardTablePagination
            disabled={refetching}
            itemLabel="promotions"
            onPageChange={setPage}
            page={pagination?.page ?? page}
            pageSize={pagination?.page_size ?? PAGE_SIZE}
            totalItems={pagination?.total ?? promotions.length}
            totalPages={pagination?.total_pages ?? 1}
          />
        </DashboardTableWrap>
      </div>
    </div>
  );
}
