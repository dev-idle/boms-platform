"use client";

import { useState } from "react";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { CustomizationSummary } from "@/components/ui/customization-summary";
import { DashboardLoadError } from "@/components/ui/dashboard-load-error";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { InlineLoadingState } from "@/components/ui/loading-state";
import {
  formatOrderStatusLabel,
  orderStatusToPillVariant,
  StatusPill,
} from "@/components/ui/status-pill";
import { isApiError } from "@/lib/errors";
import { isApplyingOrderStatus, type OrderStatus } from "@/lib/schemas/order";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";

import { usePatchStaffOrderStatus, useStaffOrder } from "../hooks";
import { staffOrderDetailBreadcrumbItems } from "../lib/staff-breadcrumbs";
import type { PatchStaffOrderStatusInput } from "../schemas";
import { StaffCancelOrderDialog } from "./staff-cancel-order-dialog";
import { StaffHandoffDialog } from "./staff-handoff-dialog";
import { StaffMovePickupDialog } from "./staff-move-pickup-dialog";
import { StaffOrderFacts } from "./staff-order-facts";
import { StaffOrderHistory } from "./staff-order-history";
import { StaffOrderTickets } from "./staff-order-tickets";
import { StaffReportIncidentDialog } from "./staff-report-incident-dialog";

/** An order staff took is paid when collected: its cash is due until then. */
const CASH_DUE_STATUSES: ReadonlySet<OrderStatus> = new Set(["confirmed", "in_production", "ready", "no_show"]);

/** An order the bakery took and has not handed over: staff may move its pickup. */
const MOVABLE_STATUSES: ReadonlySet<OrderStatus> = new Set(["pending", "confirmed", "in_production", "ready"]);

type StaffOrderDetailProps = {
  orderId: string;
};

function nextStatusActions(
  status: OrderStatus,
): Array<{ label: string; status: PatchStaffOrderStatusInput["status"] }> {
  switch (status) {
    case "pending":
      return [
        { label: "Accept order", status: "confirmed" },
        { label: "Cancel order", status: "cancelled" },
      ];
    case "confirmed":
      return [{ label: "Cancel order", status: "cancelled" }];
    case "in_production":
      return [{ label: "Cancel order", status: "cancelled" }];
    case "ready":
      return [
        { label: "Hand over", status: "fulfilled" },
        { label: "Cancel order", status: "cancelled" },
      ];
    case "no_show":
      // Missed at closing time, still handed over if its customer comes late.
      return [{ label: "Hand over", status: "fulfilled" }];
    default:
      return [];
  }
}

export function StaffOrderDetail({ orderId }: StaffOrderDetailProps) {
  const isValidId = z.uuid().safeParse(orderId).success;
  const orderQuery = useStaffOrder(orderId);
  const patchStatus = usePatchStaffOrderStatus(orderId);
  const [cancelOpen, setCancelOpen] = useState(false);
  const [handoffOpen, setHandoffOpen] = useState(false);
  const [reportOpen, setReportOpen] = useState(false);
  const [moveOpen, setMoveOpen] = useState(false);

  if (!isValidId) {
    return <p className="text-sm text-muted">This link does not lead to an order.</p>;
  }

  if (orderQuery.isPending) {
    return <InlineLoadingState />;
  }

  if (orderQuery.isError) {
    return isApiError(orderQuery.error) && orderQuery.error.status === 404 ? (
      <DashboardLoadError message="Order not found." />
    ) : (
      <DashboardLoadError message="We could not load this order." onRetry={() => void orderQuery.refetch()} />
    );
  }

  const order = orderQuery.data;
  if (!order) {
    return <p className="text-sm text-muted">Order not found.</p>;
  }

  const actions = nextStatusActions(order.status);

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <div className="dashboard-profile-form-actions">
            {actions.map((action) => (
              <Button
                key={action.status}
                disabled={patchStatus.isPending}
                type="button"
                variant={action.status === "cancelled" ? "outline" : "default"}
                onClick={() => {
                  if (action.status === "cancelled") {
                    setCancelOpen(true);
                    return;
                  }
                  if (action.status === "fulfilled") {
                    setHandoffOpen(true);
                    return;
                  }
                  patchStatus.mutate({ status: action.status });
                }}
              >
                {isApplyingOrderStatus(patchStatus, action.status)
                  ? "Updating…"
                  : action.label}
              </Button>
            ))}
            {order.pickup_at && MOVABLE_STATUSES.has(order.status) ? (
              <Button type="button" variant="outline" onClick={() => setMoveOpen(true)}>
                Move pickup
              </Button>
            ) : null}
            <Button type="button" variant="outline" onClick={() => setReportOpen(true)}>
              Report issue
            </Button>
          </div>
        }
        breadcrumbItems={staffOrderDetailBreadcrumbItems(order.code)}
        description={
          <StatusPill
            label={formatOrderStatusLabel(order.status)}
            variant={orderStatusToPillVariant(order.status)}
          />
        }
        title={order.customer.display_name ?? order.customer.email?.split("@")[0] ?? order.code}
      />

      {/* Archetype C: the order's work on the left, its facts in the right rail. */}
      <div className="dashboard-detail-grid">
        <div className="dashboard-detail-grid__main">
          <DashboardProfileSection id="staff-order-items" title="Items" variant="plain">
            <div className="dashboard-order-summary">
              <ul className="dashboard-order-line-items">
                {order.items.map((item) => (
                  <li
                    key={item.id}
                    className="dashboard-order-line-item"
                  >
                    <div className="grid gap-1">
                      <p className="font-medium text-ink">
                        {item.quantity}× {item.name}
                      </p>
                      <p className="text-muted">
                        {formatPriceCents(item.unit_price_cents)} each
                      </p>
                      {item.customization ? <CustomizationSummary customization={item.customization} /> : null}
                    </div>
                    <p className="font-medium text-tabular">
                      {formatPriceCents(item.line_total_cents)}
                    </p>
                  </li>
                ))}
              </ul>

              <div className="dashboard-order-totals">
                <div className="dashboard-order-totals-row">
                  <span className="text-muted">Subtotal</span>
                  <span className="text-tabular">{formatPriceCents(order.subtotal_cents)}</span>
                </div>
                {order.discount_cents > 0 ? (
                  <div className="dashboard-order-totals-row text-muted">
                    <span>Discount</span>
                    <span className="text-tabular">-{formatPriceCents(order.discount_cents)}</span>
                  </div>
                ) : null}
                <div className="dashboard-order-totals-row dashboard-order-totals-row--total">
                  <span>Total</span>
                  <span className="text-tabular">{formatPriceCents(order.total_cents)}</span>
                </div>
                {order.payment?.provider === "cash" ? (
                  order.payment.captured_at ? (
                    <div className="dashboard-order-totals-row text-muted">
                      <span>Paid in cash</span>
                      <span>{formatDateTime(order.payment.captured_at)}</span>
                    </div>
                  ) : CASH_DUE_STATUSES.has(order.status) ? (
                    <div className="dashboard-order-totals-row text-muted">
                      <span>Cash</span>
                      <span>Due at pickup</span>
                    </div>
                  ) : null
                ) : order.payment?.captured_at ? (
                  <div className="dashboard-order-totals-row text-muted">
                    <span>Paid with PayPal</span>
                    <span>{formatDateTime(order.payment.captured_at)}</span>
                  </div>
                ) : null}
                {order.payment?.refund_requested_at ? (
                  <div className="dashboard-order-totals-row text-muted">
                    <span>Refund to PayPal</span>
                    <span>{order.payment.refunded_at ? formatDateTime(order.payment.refunded_at) : "In progress"}</span>
                  </div>
                ) : null}
              </div>
            </div>
          </DashboardProfileSection>

          <StaffOrderTickets orderId={order.id} orderStatus={order.status} tickets={order.tickets} />

          <StaffOrderHistory timeline={order.timeline} />
        </div>
        <StaffOrderFacts order={order} />
      </div>

      <StaffCancelOrderDialog
        open={cancelOpen}
        orderId={order.id}
        refundCents={order.payment?.captured_at ? order.total_cents : null}
        onClose={() => setCancelOpen(false)}
      />

      <StaffHandoffDialog
        cashDueCents={order.channel === "online" ? null : order.total_cents}
        open={handoffOpen}
        orderCode={order.code}
        orderId={order.id}
        onClose={() => setHandoffOpen(false)}
      />

      <StaffReportIncidentDialog open={reportOpen} orderId={order.id} onClose={() => setReportOpen(false)} />

      {/* Mounted only while open: the picker reads the pickup rules and the day's slots. */}
      {moveOpen ? (
        <StaffMovePickupDialog orderCode={order.code} orderId={order.id} onClose={() => setMoveOpen(false)} />
      ) : null}
    </div>
  );
}
