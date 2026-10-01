"use client";

import { useState } from "react";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { CustomizationSummary } from "@/components/ui/customization-summary";
import { InlineLoadingState } from "@/components/ui/loading-state";
import {
  formatOrderStatusLabel,
  orderStatusToPillVariant,
  StatusPill,
} from "@/components/ui/status-pill";
import { isApiError } from "@/lib/errors";
import { formatOrderTypeLabel, isApplyingOrderStatus, type OrderStatus } from "@/lib/schemas/order";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPickupDateTime } from "@/lib/validation/pickup";
import { formatPriceCents } from "@/lib/validation/catalog";
import { formatVietnamPhone } from "@/lib/validation/phone";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";

import { usePatchStaffOrderStatus, useStaffOrder } from "../hooks";
import type { PatchStaffOrderStatusInput } from "../schemas";
import { StaffCancelOrderDialog } from "./staff-cancel-order-dialog";
import { StaffHandoffDialog } from "./staff-handoff-dialog";
import { StaffOrderHistory } from "./staff-order-history";
import { StaffOrderTickets } from "./staff-order-tickets";

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

  if (!isValidId) {
    return <p className="text-sm text-muted">Invalid order link.</p>;
  }

  if (orderQuery.isPending) {
    return <InlineLoadingState />;
  }

  if (orderQuery.isError) {
    const message =
      isApiError(orderQuery.error) && orderQuery.error.status === 404
        ? "Order not found."
        : "Failed to load order.";
    return <p className="text-sm text-error">{message}</p>;
  }

  const order = orderQuery.data;
  if (!order) {
    return <p className="text-sm text-muted">Order not found.</p>;
  }

  const actions = nextStatusActions(order.status);

  return (
    <div className="dashboard-profile-section-stack">
      <DashboardProfileSection id="staff-order-summary" title="Order summary">
        <div className="dashboard-order-summary">
          <div className="flex flex-wrap items-center gap-2">
            <p className="text-sm text-muted">
              Placed {formatDateTime(order.created_at)}
            </p>
            {order.pickup_at ? (
              <p className="text-sm text-muted">
                Pickup {formatPickupDateTime(order.pickup_at)}
              </p>
            ) : null}
            <p className="text-sm text-muted">{formatOrderTypeLabel(order.order_type)}</p>
            <StatusPill
              label={formatOrderStatusLabel(order.status)}
              variant={orderStatusToPillVariant(order.status)}
            />
          </div>
          <p className="text-order-code">{order.code}</p>
          <p className="text-sm text-muted">
            Customer:{" "}
            {[
              order.customer.display_name,
              order.customer.email,
              formatVietnamPhone(order.customer.phone),
            ]
              .filter(Boolean)
              .join(" | ")}
          </p>
          {order.discount_code_snapshot ? (
            <p className="text-sm text-muted">
              Discount code: {order.discount_code_snapshot}
            </p>
          ) : null}

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
            {order.payment?.captured_at ? (
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

          {actions.length > 0 ? (
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
            </div>
          ) : null}
        </div>
      </DashboardProfileSection>

      <StaffOrderTickets orderId={order.id} orderStatus={order.status} tickets={order.tickets} />

      <StaffOrderHistory timeline={order.timeline} />

      <StaffCancelOrderDialog
        open={cancelOpen}
        orderId={order.id}
        refundCents={order.payment?.captured_at ? order.total_cents : null}
        onClose={() => setCancelOpen(false)}
      />

      <StaffHandoffDialog
        open={handoffOpen}
        orderCode={order.code}
        orderId={order.id}
        onClose={() => setHandoffOpen(false)}
      />
    </div>
  );
}
