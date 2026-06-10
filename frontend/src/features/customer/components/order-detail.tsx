"use client";

import Link from "next/link";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useOrder } from "../hooks";

type OrderDetailProps = {
  orderId: string;
};

export function OrderDetail({ orderId }: OrderDetailProps) {
  const isValidId = z.string().uuid().safeParse(orderId).success;
  const orderQuery = useOrder(orderId);

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

  return (
    <div className="space-y-6">
      <Link href={ROUTE.orders}>
        <Button type="button" variant="outline">
          Back to orders
        </Button>
      </Link>

      <div className="rounded-lg border border-border p-4 rounded-card border bg-surface">
        <p className="text-sm text-muted">
          Placed {formatDateTime(order.created_at)} · Status: {order.status}
        </p>
        {order.discount_code_snapshot ? (
          <p className="mt-1 text-sm text-ink-2">
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
                <p className="font-medium text-ink">
                  {item.quantity}× {item.name}
                </p>
                <p className="text-muted">
                  {formatPriceCents(item.unit_price_cents)} each
                </p>
              </div>
              <p className="font-medium">
                {formatPriceCents(item.line_total_cents)}
              </p>
            </li>
          ))}
        </ul>

        <div className="mt-6 space-y-2 border-t border-border pt-4 text-sm">
          <div className="flex justify-between">
            <span className="text-ink-2">Subtotal</span>
            <span>{formatPriceCents(order.subtotal_cents)}</span>
          </div>
          {order.discount_cents > 0 ? (
            <div className="flex justify-between text-emerald-700">
              <span>Discount</span>
              <span>-{formatPriceCents(order.discount_cents)}</span>
            </div>
          ) : null}
          <div className="flex justify-between text-base font-medium">
            <span>Total</span>
            <span>{formatPriceCents(order.total_cents)}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
