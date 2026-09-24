"use client";

import { z } from "zod";
import { Button } from "@/components/ui/button";
import { InlineLoadingState } from "@/components/ui/loading-state";
import {
  formatOrderStatusLabel,
  orderStatusToPillVariant,
  StatusPill,
} from "@/components/ui/status-pill";
import { isApiError } from "@/lib/errors";
import { isApplyingOrderStatus } from "@/lib/schemas/order";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPickupDateTime } from "@/lib/validation/pickup";
import { formatPriceCents } from "@/lib/validation/catalog";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";

import { useBakerProductionOrder, usePatchBakerOrderStatus } from "../hooks";
import type { BakerOrderStatus, PatchBakerOrderStatusInput } from "../schemas";

type BakerProductionOrderDetailProps = {
  orderId: string;
};

function nextProductionActions(
  status: BakerOrderStatus,
): Array<{ label: string; status: PatchBakerOrderStatusInput["status"] }> {
  switch (status) {
    case "confirmed":
      return [{ label: "Start production", status: "in_production" }];
    case "in_production":
      return [{ label: "Mark ready", status: "ready" }];
    default:
      return [];
  }
}

export function BakerProductionOrderDetail({
  orderId,
}: BakerProductionOrderDetailProps) {
  const isValidId = z.string().uuid().safeParse(orderId).success;
  const orderQuery = useBakerProductionOrder(orderId);
  const patchStatus = usePatchBakerOrderStatus(orderId);

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
        : "Failed to load production order.";
    return <p className="text-sm text-error">{message}</p>;
  }

  const order = orderQuery.data;
  if (!order) {
    return <p className="text-sm text-muted">Order not found.</p>;
  }

  const actions = nextProductionActions(order.status);

  return (
    <DashboardProfileSection id="baker-order-summary" title="Production order">
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
          <StatusPill
            label={formatOrderStatusLabel(order.status)}
            variant={orderStatusToPillVariant(order.status)}
          />
        </div>
        <p className="text-order-code">{order.id}</p>
        <p className="text-sm text-muted">
          Customer:{" "}
          {order.customer.display_name
            ? `${order.customer.display_name} | ${order.customer.email}`
            : order.customer.email}
        </p>

        <ul className="dashboard-order-line-items">
          {order.items.map((item) => (
            <li key={item.id} className="dashboard-order-line-item">
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
        </div>

        {actions.length > 0 ? (
          <div className="dashboard-profile-form-actions">
            {actions.map((action) => (
              <Button
                key={action.status}
                disabled={patchStatus.isPending}
                type="button"
                onClick={() => patchStatus.mutate({ status: action.status })}
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
  );
}
