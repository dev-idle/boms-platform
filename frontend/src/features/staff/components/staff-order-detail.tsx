"use client";

import Link from "next/link";
import { useState } from "react";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

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
    return <p className="text-sm text-zinc-500">Invalid order link.</p>;
  }

  if (orderQuery.isPending) {
    return <p className="text-sm text-zinc-500">Loading order…</p>;
  }

  if (orderQuery.isError) {
    const message =
      isApiError(orderQuery.error) && orderQuery.error.status === 404
        ? "Order not found."
        : "Failed to load order.";
    return <p className="text-sm text-red-600">{message}</p>;
  }

  const order = orderQuery.data;
  if (!order) {
    return <p className="text-sm text-zinc-500">Order not found.</p>;
  }

  const actions = nextStatusActions(order.status);

  return (
    <div className="space-y-6">
      <Link href={ROUTE.staff.orders}>
        <Button type="button" variant="outline">
          Back to orders
        </Button>
      </Link>

      <div className="rounded-lg border border-zinc-200 p-6 dark:border-zinc-800">
        <p className="text-sm text-zinc-500">
          Placed {formatDateTime(order.created_at)} · Status: {order.status}
        </p>
        <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
          Customer:{" "}
          {order.customer.display_name
            ? `${order.customer.display_name} · `
            : ""}
          {order.customer.email}
        </p>
        {order.discount_code_snapshot ? (
          <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
            Discount code: {order.discount_code_snapshot}
          </p>
        ) : null}

        <ul className="mt-6 space-y-3">
          {order.items.map((item) => (
            <li
              key={item.id}
              className="flex items-start justify-between gap-4 text-sm"
            >
              <div>
                <p className="font-medium text-zinc-900 dark:text-zinc-50">
                  {item.quantity}× {item.name}
                </p>
                <p className="text-zinc-500">
                  {formatPriceCents(item.unit_price_cents)} each
                </p>
              </div>
              <p className="font-medium">
                {formatPriceCents(item.line_total_cents)}
              </p>
            </li>
          ))}
        </ul>

        <div className="mt-6 space-y-2 border-t border-zinc-200 pt-4 text-sm dark:border-zinc-800">
          <div className="flex justify-between">
            <span className="text-zinc-600 dark:text-zinc-400">Subtotal</span>
            <span>{formatPriceCents(order.subtotal_cents)}</span>
          </div>
          {order.discount_cents > 0 ? (
            <div className="flex justify-between text-emerald-700 dark:text-emerald-400">
              <span>Discount</span>
              <span>-{formatPriceCents(order.discount_cents)}</span>
            </div>
          ) : null}
          <div className="flex justify-between text-base font-semibold">
            <span>Total</span>
            <span>{formatPriceCents(order.total_cents)}</span>
          </div>
        </div>

        {actions.length > 0 ? (
          <div className="mt-6 flex flex-wrap gap-2">
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

      <ConfirmDialog
        cancelLabel="Keep order"
        confirmLabel="Cancel order"
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
    </div>
  );
}
