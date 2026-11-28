"use client";

import { useState } from "react";

import { DashboardTableActionButton } from "@/components/ui/dashboard-table-action-button";
import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { DashboardTableRowActions } from "@/components/ui/dashboard-table-actions";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { DateInput } from "@/components/ui/date-input";
import { DashboardStepGroup } from "@/components/ui/dashboard-step-group";
import {
  formatOrderStatusLabel,
  orderStatusToPillVariant,
  StatusPill,
} from "@/components/ui/status-pill";
import { ROUTE } from "@/constants/routes";
import { useNow } from "@/lib/hooks/use-now";
import { keepPageInRange } from "@/lib/pagination/dashboard-pagination";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import type { OrderStatus } from "@/lib/schemas/order";
import { formatTicketItems } from "@/lib/schemas/ticket";
import { plural } from "@/lib/utils";
import { bakeryDayOf, formatPickupWallTime, shiftDay } from "@/lib/validation/pickup";

import { useStaffPickups } from "../hooks";
import { counterReadyBlockedReason, groupPickupsBySlot, isLatePickup } from "../lib/pickup-schedule";
import type { StaffPickup } from "../schemas";
import { StaffCounterReadyDialog } from "./staff-counter-ready-dialog";
import { StaffCustomerCell } from "./staff-customer-cell";
import { StaffHandoffDialog } from "./staff-handoff-dialog";

const COLUMN_COUNT = 5;
const MINUTE_MS = 60_000;
const PAGE_SIZE = 50;

/** Why the counter cannot hand this order over, or undefined when it can. */
function handoffBlockedReason(status: OrderStatus): string | undefined {
  switch (status) {
    case "ready":
      return undefined;
    case "no_show":
      // A missed order is still handed over if its customer comes late.
      return undefined;
    case "fulfilled":
      return "Already picked up";
    default:
      return "Not ready yet";
  }
}

/**
 * The counter's day: every pickup by time, late ones flagged. The counter gets
 * its own part of each order ready here — drinks and items from the case — and
 * hands over the orders every station has finished.
 */
export function StaffPickupSchedule() {
  const now = useNow(MINUTE_MS);
  const today = bakeryDayOf(now);
  // Unless staff pick a day, the schedule is today's, and turns over at midnight.
  const [chosenDay, setChosenDay] = useState<string | null>(null);
  const day = chosenDay ?? today;
  // A page belongs to its day: another day, picked or come at midnight, starts on page 1.
  const [paging, setPaging] = useState({ day, page: 1 });
  const page = paging.day === day ? paging.page : 1;
  const [handoff, setHandoff] = useState<StaffPickup | null>(null);
  const [handoffOpen, setHandoffOpen] = useState(false);
  const [preparing, setPreparing] = useState<StaffPickup | null>(null);
  const pickupsQuery = useStaffPickups({ date: day, page, page_size: PAGE_SIZE });
  const { initialLoading, refetching } = getQuerySurface(pickupsQuery);
  const pickups = pickupsQuery.data?.pickups ?? [];
  const slotMinutes = pickupsQuery.data?.slot_minutes ?? 0;
  const pagination = pickupsQuery.data?.pagination;
  keepPageInRange(page, (next) => setPaging({ day, page: next }), pagination?.total_pages);
  const showState = initialLoading || pickupsQuery.isError || pickups.length === 0;

  function showDay(next: string): void {
    setChosenDay(next === today ? null : next);
  }

  return (
    <div className="dashboard-page-body">
      <div className="db-table-filters">
        <div className="db-table-day">
          <DateInput id="staff-pickups-day" label="Pickup day" placeholder="Choose a day" required value={day} onChange={showDay} />
        </div>
        <DashboardStepGroup
          aria-label="Step the pickup day"
          steps={[
            { label: "Previous day", direction: "previous", onClick: () => showDay(shiftDay(day, -1)) },
            { label: "Today", blockedReason: day === today ? "Already showing today" : undefined, onClick: () => showDay(today) },
            { label: "Next day", direction: "next", onClick: () => showDay(shiftDay(day, 1)) },
          ]}
        />
      </div>

      <DashboardTableWrap refetching={refetching}>
        <table className="db-table db-table--pickups">
          <colgroup>
            <col className="db-table-col-code" />
            <col />
            <col className="db-table-col-counter" />
            <col className="db-table-col-status" />
            <col className="db-table-col-actions" />
          </colgroup>
          <thead>
            <tr>
              <th scope="col">Order</th>
              <th scope="col">Customer</th>
              <th scope="col">Counter</th>
              <th scope="col" className="db-table-status">Status</th>
              <th scope="col" className="db-table-detail">Actions</th>
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
                onRetry={() => void pickupsQuery.refetch()}
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
                        <StaffCustomerCell channel={pickup.channel} customer={pickup.customer} />
                      </td>
                      <td>
                        {pickup.counter_ticket ? (
                          formatTicketItems(pickup.counter_ticket.items)
                        ) : (
                          <span className="text-muted">Kitchen only</span>
                        )}
                      </td>
                      <td className="db-table-status">
                        <StatusPill
                          label={formatOrderStatusLabel(pickup.status)}
                          variant={orderStatusToPillVariant(pickup.status)}
                        />
                      </td>
                      <td className="db-table-detail">
                        <DashboardTableRowActions>
                          <DashboardTableActionButton
                            blockedReason={counterReadyBlockedReason(pickup)}
                            label={`Mark the counter part of order ${pickup.code} ready`}
                            onClick={() => setPreparing(pickup)}
                            text="Ready"
                            tone="accent"
                          />
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
          onPageChange={(next) => setPaging({ day, page: next })}
          page={pagination?.page ?? page}
          pageSize={pagination?.page_size ?? PAGE_SIZE}
          totalItems={pagination?.total ?? pickups.length}
          totalPages={pagination?.total_pages ?? 1}
        />
      </DashboardTableWrap>

      <StaffCounterReadyDialog pickup={preparing} onClose={() => setPreparing(null)} />

      {/* Stays mounted once opened, so closing it hands focus back to the row. */}
      {handoff ? (
        <StaffHandoffDialog
          cashDueCents={handoff.channel === "online" ? null : handoff.total_cents}
          open={handoffOpen}
          orderCode={handoff.code}
          orderId={handoff.id}
          onClose={() => setHandoffOpen(false)}
        />
      ) : null}
    </div>
  );
}
