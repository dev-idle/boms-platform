"use client";

import { useState } from "react";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { DashboardTableActionButton } from "@/components/ui/dashboard-table-action-button";
import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { DashboardTableRowActions } from "@/components/ui/dashboard-table-actions";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { Input } from "@/components/ui/input";
import {
  formatOrderStatusLabel,
  orderStatusToPillVariant,
  StatusPill,
} from "@/components/ui/status-pill";
import { ROUTE } from "@/constants/routes";
import { useNow } from "@/lib/hooks/use-now";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import type { OrderStatus } from "@/lib/schemas/order";
import { bakeryDayOf, formatPickupWallTime } from "@/lib/validation/pickup";

import { useStaffPickups } from "../hooks";
import { groupPickupsBySlot, isLatePickup, shiftDay } from "../lib/pickup-schedule";
import type { StaffPickup } from "../schemas";
import { StaffHandoffDialog } from "./staff-handoff-dialog";

const COLUMN_COUNT = 5;
const MINUTE_MS = 60_000;
const PAGE_SIZE = 50;

/** Why the counter cannot hand this order over, or undefined when it can. */
function handoffBlockedReason(status: OrderStatus): string | undefined {
  switch (status) {
    case "ready":
      return undefined;
    case "fulfilled":
      return "Already picked up";
    case "no_show":
      return "Marked as not collected";
    default:
      return "Not ready yet";
  }
}

function plural(count: number, noun: string): string {
  return `${count} ${noun}${count === 1 ? "" : "s"}`;
}

/** The counter's day: every pickup by time, late ones flagged, ready ones handed over here. */
export function StaffPickupSchedule() {
  const now = useNow(MINUTE_MS);
  const today = bakeryDayOf(now);
  const [day, setDay] = useState(today);
  const [page, setPage] = useState(1);
  const [handoff, setHandoff] = useState<StaffPickup | null>(null);
  const [handoffOpen, setHandoffOpen] = useState(false);
  const pickupsQuery = useStaffPickups({ date: day, page, page_size: PAGE_SIZE });
  const { initialLoading, refetching } = getQuerySurface(pickupsQuery);
  const pickups = pickupsQuery.data?.pickups ?? [];
  const slotMinutes = pickupsQuery.data?.slot_minutes ?? 0;
  const pagination = pickupsQuery.data?.pagination;
  const showState = initialLoading || pickupsQuery.isError || pickups.length === 0;

  function showDay(next: string): void {
    setDay(next);
    setPage(1);
  }

  return (
    <div className="dashboard-page-body">
      <div className="db-table-filters">
        <div className="db-table-day">
          <Input
            aria-label="Pickup day"
            type="date"
            value={day}
            variant="inline"
            onChange={(event) => {
              const parsed = z.iso.date().safeParse(event.target.value);
              if (parsed.success) {
                showDay(parsed.data);
              }
            }}
          />
          <Button size="sm" type="button" variant="outline" onClick={() => showDay(shiftDay(day, -1))}>
            Previous day
          </Button>
          <Button
            disabled={day === today}
            size="sm"
            title={day === today ? "Already showing today" : undefined}
            type="button"
            variant="outline"
            onClick={() => showDay(today)}
          >
            Today
          </Button>
          <Button size="sm" type="button" variant="outline" onClick={() => showDay(shiftDay(day, 1))}>
            Next day
          </Button>
        </div>
      </div>

      <DashboardTableWrap refetching={refetching}>
        <table className="db-table db-table--pickups">
          <colgroup>
            <col className="db-table-col-code" />
            <col />
            <col className="db-table-col-number" />
            <col className="db-table-col-status" />
            <col className="db-table-col-actions" />
          </colgroup>
          <thead>
            <tr>
              <th>Order</th>
              <th>Customer</th>
              <th className="db-table-num">Items</th>
              <th className="db-table-status">Status</th>
              <th className="db-table-detail">Actions</th>
            </tr>
          </thead>
          {showState ? (
            <tbody>
              <DashboardTableStateRows
                columnCount={COLUMN_COUNT}
                emptyMessage="No pickups on this day."
                entityLabel="pickups"
                isEmpty={pickups.length === 0}
                isError={pickupsQuery.isError}
                initialLoading={initialLoading}
              />
            </tbody>
          ) : (
            groupPickupsBySlot(pickups).map((slot) => {
              const ready = slot.pickups.filter((pickup) => pickup.status === "ready").length;
              const late = slot.pickups.filter((pickup) => isLatePickup(pickup, slotMinutes, now)).length;
              return (
                <tbody key={slot.at}>
                  <tr className="db-table-slot">
                    <th colSpan={COLUMN_COUNT} scope="rowgroup">
                      <span className="db-table-slot__head">
                        <span className="db-table-slot__time">{formatPickupWallTime(slot.at)}</span>
                        <span className="db-table-slot__meta">
                          {plural(slot.pickups.length, "pickup")}
                          {ready > 0 ? ` · ${ready} ready` : null}
                        </span>
                        {late > 0 ? <span className="db-table-slot__late">{late} late</span> : null}
                      </span>
                    </th>
                  </tr>
                  {slot.pickups.map((pickup) => (
                    <tr key={pickup.id}>
                      <td className="whitespace-nowrap">
                        <span className="text-order-code">{pickup.code}</span>
                      </td>
                      <td>
                        <div className="db-table-stacked-cell min-w-0">
                          <span className="db-table-cell-primary truncate">
                            {pickup.customer.display_name ?? pickup.customer.email}
                          </span>
                          {pickup.customer.display_name ? (
                            <span className="truncate text-caption-dashboard text-muted">
                              {pickup.customer.email}
                            </span>
                          ) : null}
                        </div>
                      </td>
                      <td className="db-table-num">{pickup.item_count}</td>
                      <td className="db-table-status">
                        <StatusPill
                          label={formatOrderStatusLabel(pickup.status)}
                          variant={orderStatusToPillVariant(pickup.status)}
                        />
                      </td>
                      <td className="db-table-detail">
                        <DashboardTableRowActions>
                          <DashboardTableActionButton
                            blockedReason={handoffBlockedReason(pickup.status)}
                            label={`Hand over order ${pickup.code}`}
                            onClick={() => {
                              setHandoff(pickup);
                              setHandoffOpen(true);
                            }}
                            text="Hand over"
                            tone="accent"
                          />
                          <DashboardTableActionLink
                            href={ROUTE.staff.orderDetail(pickup.id)}
                            label={`Open order ${pickup.code}`}
                            showArrow
                            text="Order"
                          />
                        </DashboardTableRowActions>
                      </td>
                    </tr>
                  ))}
                </tbody>
              );
            })
          )}
        </table>
        <DashboardTablePagination
          disabled={refetching}
          itemLabel="pickups"
          onPageChange={setPage}
          page={pagination?.page ?? page}
          pageSize={pagination?.page_size ?? PAGE_SIZE}
          totalItems={pagination?.total ?? pickups.length}
          totalPages={pagination?.total_pages ?? 1}
        />
      </DashboardTableWrap>

      {/* Stays mounted once opened, so closing it hands focus back to the row. */}
      {handoff ? (
        <StaffHandoffDialog
          open={handoffOpen}
          orderCode={handoff.code}
          orderId={handoff.id}
          onClose={() => setHandoffOpen(false)}
        />
      ) : null}
    </div>
  );
}
