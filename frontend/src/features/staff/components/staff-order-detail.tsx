"use client";

import { useState } from "react";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import {
  formatOrderStatusLabel,
  orderStatusToPillVariant,
  StatusPill,
} from "@/components/ui/status-pill";
import { isApiError } from "@/lib/errors";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";
import { DashboardProfileSection } from "@/features/user";

import { usePatchStaffOrderStatus, useStaffOrder } from "../hooks";
import type { OrderStatus, PatchStaffOrderStatusInput } from "../schemas";

type StaffOrderDetailProps = {
  orderId: string;
};

function nextStatusActions(
  status: OrderStatus,
): Array<{ label: string; status: PatchStaffOrderStatusInput["status"] }> {
  switch (status) {
    case "pending":
      return [
        { label: "Confirm order", status: "confirmed" },
        { label: "Cancel order", status: "cancelled" },
      ];
    case "confirmed":
      return [
        { label: "Mark fulfilled", status: "fulfilled" },
        { label: "Cancel order", status: "cancelled" },
      ];
    default:
      return [];
  }
}

export function StaffOrderDetail({ orderId }: StaffOrderDetailProps) {
  const isValidId = z.string().uuid().safeParse(orderId).success;
  const orderQuery = useStaffOrder(orderId);
  const patchStatus = usePatchStaffOrderStatus(orderId);
  const [cancelOpen, setCancelOpen] = useState(false);

  if (!isValidId) {
    return <p className="text-sm text-muted">Invalid order link.</p>;
  }

  if (orderQuery.isPending) {
    return <p className="text-sm text-muted">Loading order…</p>;
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
    <>
      <DashboardProfileSection id="staff-order-summary" title="Order summary">
        <div className="dashboard-order-summary">
          <div className="flex flex-wrap items-center gap-2">
            <p className="text-sm text-muted">
              Placed {formatDateTime(order.created_at)}
            </p>
            <StatusPill
              label={formatOrderStatusLabel(order.status)}
              variant={orderStatusToPillVariant(order.status)}
            />
          </div>
          <p className="text-order-code">{order.id}</p>
          <p className="text-sm text-ink-2">
            Customer:{" "}
            {order.customer.display_name
              ? `${order.customer.display_name} | ${order.customer.email}`
              : order.customer.email}
          </p>
          {order.discount_code_snapshot ? (
            <p className="text-sm text-ink-2">
              Discount code: {order.discount_code_snapshot}
            </p>
          ) : null}

          <ul className="dashboard-order-line-items">
            {order.items.map((item) => (
              <li
                key={item.id}
                className="dashboard-order-line-item"
              >
                <div>
                  <p className="font-medium text-ink">
                    {item.quantity}× {item.name}
                  </p>
                  <p className="text-muted">
                    {formatPriceCents(item.unit_price_cents)} each
                  </p>
                </div>
                <p className="font-medium text-tabular">
                  {formatPriceCents(item.line_total_cents)}
                </p>
              </li>
            ))}
          </ul>

          <div className="dashboard-order-totals">
            <div className="dashboard-order-totals-row">
              <span className="text-ink-2">Subtotal</span>
              <span className="text-tabular">{formatPriceCents(order.subtotal_cents)}</span>
            </div>
            {order.discount_cents > 0 ? (
              <div className="dashboard-order-totals-row text-ink-2">
                <span>Discount</span>
                <span className="text-tabular">-{formatPriceCents(order.discount_cents)}</span>
              </div>
            ) : null}
            <div className="dashboard-order-totals-row dashboard-order-totals-row--total">
              <span>Total</span>
              <span className="text-tabular">{formatPriceCents(order.total_cents)}</span>
            </div>
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
                    patchStatus.mutate({ status: action.status });
                  }}
                >
                  {patchStatus.isPending ? "Updating…" : action.label}
                </Button>
              ))}
            </div>
          ) : null}
        </div>
      </DashboardProfileSection>

      <ConfirmDialog
        cancelLabel="Keep order"
        confirmLabel="Cancel order"
        confirmVariant="destructive"
        description="The customer will see this order as cancelled."
        isPending={patchStatus.isPending}
        onCancel={() => setCancelOpen(false)}
        onConfirm={() =>
          patchStatus.mutate(
            { status: "cancelled" },
            { onSuccess: () => setCancelOpen(false) },
          )
        }
        open={cancelOpen}
        title="Cancel this order?"
      />
    </>
  );
}
